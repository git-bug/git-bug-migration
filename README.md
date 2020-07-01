Packages Needed:
 - Identity
 - Repository
 - Entity
 - Util

Methodology:  
This tool is designed to import a repository using an outdated version of git-bug and update it to be compatible with a
newer version. This allows for breaking changes to git-bug without completely crashing the repositories of the user.

This tool will contain all versions of git-bug with breaking changes. It will make any necessary edits to the objects to
convert them between the two versions. If the repository needs to jump between multiple versions, it will then complete
the steps necessary to jump to the next version and so on until the target version is reached.

Sections:
 - Identity
 - OperationPack
 - Config
 
Todo for migration tool:  
 - Remove legacy author and create new identity out of it -- forwards compatible
 - https://github.com/MichaelMure/git-bug-migration/migration1/pull/411: identity.Version need to store one time + Lamport time for each
 other Entity (Bug, config, PR ...) instead of a single one for Bug at the moment.
 - At the moment bridge credentials are stored in the global git config. In the future it could be stored in a real
credential store. Migrating that would be nice.
 - AvatarURL in the identities reference an external images. Eventually storing all that in git itself would be nice.
Migration would be downloading all that for offline use and rewriting those references.

Todo for git-bug:  
- Bump the OperationPack format version to 2 so that the legacy code can be removed. We don't need it to still be able
to read the data, but when we remove the legacy code we need to be able to detect that we can't read that anymore.