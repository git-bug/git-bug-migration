// Package identity contains the identity data model and low-level related functions
package identity

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/MichaelMure/git-bug-migration/migration3/before/entity"
	"github.com/MichaelMure/git-bug-migration/migration3/before/repository"
	"github.com/MichaelMure/git-bug-migration/migration3/before/util/lamport"
	"github.com/MichaelMure/git-bug-migration/migration3/before/util/timestamp"
)

const identityRefPattern = "refs/identities/"
const identityRemoteRefPattern = "refs/remotes/%s/identities/"
const versionEntryName = "version"
const identityConfigKey = "git-bug.identity"

var ErrNonFastForwardMerge = errors.New("non fast-forward identity merge")
var ErrNoIdentitySet = errors.New("No identity is set.\n" +
	"To interact with bugs, an identity first needs to be created using " +
	"\"git bug user create\"")
var ErrMultipleIdentitiesSet = errors.New("multiple user identities set")

func NewErrMultipleMatchIdentity(matching []entity.Id) *entity.ErrMultipleMatch {
	return entity.NewErrMultipleMatch("identity", matching)
}

var _ Interface = &Identity{}
var _ entity.Interface = &Identity{}

type Identity struct {
	// Id used as unique identifier
	id entity.Id

	// all the successive version of the identity
	versions []*Version

	// not serialized
	lastCommit repository.Hash
}

func NewIdentity(name string, email string) *Identity {
	return &Identity{
		id: entity.UnsetId,
		versions: []*Version{
			{
				name:  name,
				email: email,
				nonce: makeNonce(20),
			},
		},
	}
}

func NewIdentityFull(name string, email string, login string, avatarUrl string) *Identity {
	return &Identity{
		id: entity.UnsetId,
		versions: []*Version{
			{
				name:      name,
				email:     email,
				login:     login,
				avatarURL: avatarUrl,
				nonce:     makeNonce(20),
			},
		},
	}
}

// NewFromGitUser will query the repository for user detail and
// build the corresponding Identity
func NewFromGitUser(repo repository.Repo) (*Identity, error) {
	name, err := repo.GetUserName()
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("user name is not configured in git yet. Please use `git config --global user.name \"John Doe\"`")
	}

	email, err := repo.GetUserEmail()
	if err != nil {
		return nil, err
	}
	if email == "" {
		return nil, errors.New("user name is not configured in git yet. Please use `git config --global user.email johndoe@example.com`")
	}

	return NewIdentity(name, email), nil
}

// MarshalJSON will only serialize the id
func (i *Identity) MarshalJSON() ([]byte, error) {
	return json.Marshal(&IdentityStub{
		id: i.id,
	})
}

// UnmarshalJSON will only read the id
// Users of this package are expected to run Load() to load
// the remaining data from the identities data in git.
func (i *Identity) UnmarshalJSON(data []byte) error {
	panic("identity should be loaded with identity.UnmarshalJSON")
}

// ReadLocal load a local Identity from the identities data available in git
func ReadLocal(repo repository.Repo, id entity.Id) (*Identity, error) {
	ref := fmt.Sprintf("%s%s", identityRefPattern, id)
	return read(repo, ref)
}

// ReadRemote load a remote Identity from the identities data available in git
func ReadRemote(repo repository.Repo, remote string, id string) (*Identity, error) {
	ref := fmt.Sprintf(identityRemoteRefPattern, remote) + id
	return read(repo, ref)
}

// read will load and parse an identity from git
func read(repo repository.Repo, ref string) (*Identity, error) {
	refSplit := strings.Split(ref, "/")
	id := entity.Id(refSplit[len(refSplit)-1])

	if err := id.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid ref")
	}

	hashes, err := repo.ListCommits(ref)

	// TODO: this is not perfect, it might be a command invoke error
	if err != nil {
		return nil, ErrIdentityNotExist
	}

	i := &Identity{
		id: id,
	}

	for _, hash := range hashes {
		entries, err := repo.ReadTree(hash)
		if err != nil {
			return nil, errors.Wrap(err, "can't list git tree entries")
		}

		if len(entries) != 1 {
			return nil, fmt.Errorf("invalid identity data at hash %s", hash)
		}

		entry := entries[0]

		if entry.Name != versionEntryName {
			return nil, fmt.Errorf("invalid identity data at hash %s", hash)
		}

		data, err := repo.ReadData(entry.Hash)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read git blob data")
		}

		var version Version
		err = json.Unmarshal(data, &version)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode Identity version json %s", hash)
		}

		// tag the version with the commit hash
		version.commitHash = hash
		i.lastCommit = hash

		i.versions = append(i.versions, &version)
	}

	return i, nil
}

// ListLocalIds list all the available local identity ids
func ListLocalIds(repo repository.Repo) ([]entity.Id, error) {
	refs, err := repo.ListRefs(identityRefPattern)
	if err != nil {
		return nil, err
	}

	return entity.RefsToIds(refs), nil
}

// RemoveIdentity will remove a local identity from its entity.Id
func RemoveIdentity(repo repository.ClockedRepo, id entity.Id) error {
	var fullMatches []string

	refs, err := repo.ListRefs(identityRefPattern + id.String())
	if err != nil {
		return err
	}
	if len(refs) > 1 {
		return NewErrMultipleMatchIdentity(entity.RefsToIds(refs))
	}
	if len(refs) == 1 {
		// we have the identity locally
		fullMatches = append(fullMatches, refs[0])
	}

	remotes, err := repo.GetRemotes()
	if err != nil {
		return err
	}

	for remote := range remotes {
		remotePrefix := fmt.Sprintf(identityRemoteRefPattern+id.String(), remote)
		remoteRefs, err := repo.ListRefs(remotePrefix)
		if err != nil {
			return err
		}
		if len(remoteRefs) > 1 {
			return NewErrMultipleMatchIdentity(entity.RefsToIds(refs))
		}
		if len(remoteRefs) == 1 {
			// found the identity in a remote
			fullMatches = append(fullMatches, remoteRefs[0])
		}
	}

	if len(fullMatches) == 0 {
		return ErrIdentityNotExist
	}

	for _, ref := range fullMatches {
		err = repo.RemoveRef(ref)
		if err != nil {
			return err
		}
	}

	return nil
}

type StreamedIdentity struct {
	Identity *Identity
	Err      error
}

// ReadAllLocal read and parse all local Identity
func ReadAllLocal(repo repository.ClockedRepo) <-chan StreamedIdentity {
	return readAll(repo, identityRefPattern)
}

// ReadAllRemote read and parse all remote Identity for a given remote
func ReadAllRemote(repo repository.ClockedRepo, remote string) <-chan StreamedIdentity {
	refPrefix := fmt.Sprintf(identityRemoteRefPattern, remote)
	return readAll(repo, refPrefix)
}

// readAll read and parse all available bug with a given ref prefix
func readAll(repo repository.ClockedRepo, refPrefix string) <-chan StreamedIdentity {
	out := make(chan StreamedIdentity)

	go func() {
		defer close(out)

		refs, err := repo.ListRefs(refPrefix)
		if err != nil {
			out <- StreamedIdentity{Err: err}
			return
		}

		for _, ref := range refs {
			b, err := read(repo, ref)

			if err != nil {
				out <- StreamedIdentity{Err: err}
				return
			}

			out <- StreamedIdentity{Identity: b}
		}
	}()

	return out
}

type Mutator struct {
	Name      string
	Login     string
	Email     string
	AvatarUrl string
	Keys      []*Key
}

// Mutate allow to create a new version of the Identity in one go
func (i *Identity) Mutate(f func(orig Mutator) Mutator) {
	orig := Mutator{
		Name:      i.Name(),
		Email:     i.Email(),
		Login:     i.Login(),
		AvatarUrl: i.AvatarUrl(),
		Keys:      i.Keys(),
	}
	mutated := f(orig)
	if reflect.DeepEqual(orig, mutated) {
		return
	}
	i.versions = append(i.versions, &Version{
		name:      mutated.Name,
		email:     mutated.Email,
		login:     mutated.Login,
		avatarURL: mutated.AvatarUrl,
		keys:      mutated.Keys,
	})
}

// Write the identity into the Repository. In particular, this ensure that
// the Id is properly set.
func (i *Identity) Commit(repo repository.ClockedRepo) error {
	// Todo: check for mismatch between memory and commit data

	if !i.NeedCommit() {
		return fmt.Errorf("can't commit an identity with no pending version")
	}

	if err := i.Validate(); err != nil {
		return errors.Wrap(err, "can't commit an identity with invalid data")
	}

	for _, v := range i.versions {
		if v.commitHash != "" {
			i.lastCommit = v.commitHash
			// ignore already commit versions
			continue
		}

		// get the times where new versions starts to be valid
		// TODO: instead of this hardcoded clock for bugs only, this need to be
		// a vector of edit clock, one for each entity (bug, PR, config ..)
		bugEditClock, err := repo.GetOrCreateClock("bug-edit")
		if err != nil {
			return err
		}

		v.time = bugEditClock.Time()
		v.unixTime = time.Now().Unix()

		blobHash, err := v.Write(repo)
		if err != nil {
			return err
		}

		// Make a git tree referencing the blob
		tree := []repository.TreeEntry{
			{ObjectType: repository.Blob, Hash: blobHash, Name: versionEntryName},
		}

		treeHash, err := repo.StoreTree(tree)
		if err != nil {
			return err
		}

		var commitHash repository.Hash
		if i.lastCommit != "" {
			commitHash, err = repo.StoreCommitWithParent(treeHash, i.lastCommit)
		} else {
			commitHash, err = repo.StoreCommit(treeHash)
		}

		if err != nil {
			return err
		}

		i.lastCommit = commitHash
		v.commitHash = commitHash

		// if it was the first commit, use the commit hash as the Identity id
		if i.id == "" || i.id == entity.UnsetId {
			i.id = entity.Id(commitHash)
		}
	}

	if i.id == "" {
		panic("identity with no id")
	}

	ref := fmt.Sprintf("%s%s", identityRefPattern, i.id)
	err := repo.UpdateRef(ref, i.lastCommit)

	if err != nil {
		return err
	}

	return nil
}

func (i *Identity) CommitAsNeeded(repo repository.ClockedRepo) error {
	if !i.NeedCommit() {
		return nil
	}
	return i.Commit(repo)
}

func (i *Identity) NeedCommit() bool {
	for _, v := range i.versions {
		if v.commitHash == "" {
			return true
		}
	}

	return false
}

// Merge will merge a different version of the same Identity
//
// To make sure that an Identity history can't be altered, a strict fast-forward
// only policy is applied here. As an Identity should be tied to a single user, this
// should work in practice but it does leave a possibility that a user would edit his
// Identity from two different repo concurrently and push the changes in a non-centralized
// network of repositories. In this case, it would result in some of the repo accepting one
// version and some other accepting another, preventing the network in general to converge
// to the same result. This would create a sort of partition of the network, and manual
// cleaning would be required.
//
// An alternative approach would be to have a determinist rebase:
// - any commits present in both local and remote version would be kept, never changed.
// - newer commits would be merged in a linear chain of commits, ordered based on the
//   Lamport time
//
// However, this approach leave the possibility, in the case of a compromised crypto keys,
// of forging a new version with a bogus Lamport time to be inserted before a legit version,
// invalidating the correct version and hijacking the Identity. There would only be a short
// period of time where this would be possible (before the network converge) but I'm not
// confident enough to implement that. I choose the strict fast-forward only approach,
// despite it's potential problem with two different version as mentioned above.
func (i *Identity) Merge(repo repository.Repo, other *Identity) (bool, error) {
	if i.id != other.id {
		return false, errors.New("merging unrelated identities is not supported")
	}

	if i.lastCommit == "" || other.lastCommit == "" {
		return false, errors.New("can't merge identities that has never been stored")
	}

	modified := false
	for j, otherVersion := range other.versions {
		// if there is more version in other, take them
		if len(i.versions) == j {
			i.versions = append(i.versions, otherVersion)
			i.lastCommit = otherVersion.commitHash
			modified = true
		}

		// we have a non fast-forward merge.
		// as explained in the doc above, refusing to merge
		if i.versions[j].commitHash != otherVersion.commitHash {
			return false, ErrNonFastForwardMerge
		}
	}

	if modified {
		err := repo.UpdateRef(identityRefPattern+i.id.String(), i.lastCommit)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// Validate check if the Identity data is valid
func (i *Identity) Validate() error {
	lastTime := lamport.Time(0)

	if len(i.versions) == 0 {
		return fmt.Errorf("no version")
	}

	for _, v := range i.versions {
		if err := v.Validate(); err != nil {
			return err
		}

		if v.commitHash != "" && v.time < lastTime {
			return fmt.Errorf("non-chronological version (%d --> %d)", lastTime, v.time)
		}

		lastTime = v.time
	}

	// The identity Id should be the hash of the first commit
	if i.versions[0].commitHash != "" && string(i.versions[0].commitHash) != i.id.String() {
		return fmt.Errorf("identity id should be the first commit hash")
	}

	return nil
}

func (i *Identity) lastVersion() *Version {
	if len(i.versions) <= 0 {
		panic("no version at all")
	}

	return i.versions[len(i.versions)-1]
}

// Id return the Identity identifier
func (i *Identity) Id() entity.Id {
	if i.id == "" || i.id == entity.UnsetId {
		// simply panic as it would be a coding error
		// (using an id of an identity not stored yet)
		panic("no id yet")
	}
	return i.id
}

// Name return the last version of the name
func (i *Identity) Name() string {
	return i.lastVersion().name
}

// Email return the last version of the email
func (i *Identity) Email() string {
	return i.lastVersion().email
}

// Login return the last version of the login
func (i *Identity) Login() string {
	return i.lastVersion().login
}

// AvatarUrl return the last version of the Avatar URL
func (i *Identity) AvatarUrl() string {
	return i.lastVersion().avatarURL
}

// Keys return the last version of the valid keys
func (i *Identity) Keys() []*Key {
	return i.lastVersion().keys
}

// ValidKeysAtTime return the set of keys valid at a given lamport time
func (i *Identity) ValidKeysAtTime(time lamport.Time) []*Key {
	var result []*Key

	for _, v := range i.versions {
		if v.time > time {
			return result
		}

		result = v.keys
	}

	return result
}

// DisplayName return a non-empty string to display, representing the
// identity, based on the non-empty values.
func (i *Identity) DisplayName() string {
	switch {
	case i.Name() == "" && i.Login() != "":
		return i.Login()
	case i.Name() != "" && i.Login() == "":
		return i.Name()
	case i.Name() != "" && i.Login() != "":
		return fmt.Sprintf("%s (%s)", i.Name(), i.Login())
	}

	panic("invalid person data")
}

// IsProtected return true if the chain of git commits started to be signed.
// If that's the case, only signed commit with a valid key for this identity can be added.
func (i *Identity) IsProtected() bool {
	// Todo
	return false
}

// LastModificationLamportTime return the Lamport time at which the last version of the identity became valid.
func (i *Identity) LastModificationLamport() lamport.Time {
	return i.lastVersion().time
}

// LastModification return the timestamp at which the last version of the identity became valid.
func (i *Identity) LastModification() timestamp.Timestamp {
	return timestamp.Timestamp(i.lastVersion().unixTime)
}

// SetMetadata store arbitrary metadata along the last not-commit Version.
// If the Version has been commit to git already, a new identical version is added and will need to be
// commit.
func (i *Identity) SetMetadata(key string, value string) {
	if i.lastVersion().commitHash != "" {
		i.versions = append(i.versions, i.lastVersion().Clone())
	}
	i.lastVersion().SetMetadata(key, value)
}

// ImmutableMetadata return all metadata for this Identity, accumulated from each Version.
// If multiple value are found, the first defined takes precedence.
func (i *Identity) ImmutableMetadata() map[string]string {
	metadata := make(map[string]string)

	for _, version := range i.versions {
		for key, value := range version.metadata {
			if _, has := metadata[key]; !has {
				metadata[key] = value
			}
		}
	}

	return metadata
}

// MutableMetadata return all metadata for this Identity, accumulated from each Version.
// If multiple value are found, the last defined takes precedence.
func (i *Identity) MutableMetadata() map[string]string {
	metadata := make(map[string]string)

	for _, version := range i.versions {
		for key, value := range version.metadata {
			metadata[key] = value
		}
	}

	return metadata
}

// addVersionForTest add a new version to the identity
// Only for testing !
func (i *Identity) addVersionForTest(version *Version) {
	i.versions = append(i.versions, version)
}