# Chexum Extension Tutorials

## Adding a New Hashing Algorithm

To add a new algorithm (e.g., `SHA3-256`):

1.  **Define the Constant**: Add the algorithm name to `internal/hash/hash.go`:
    ```go
    const AlgorithmSHA3_256 = "sha3-256"
    ```
2.  **Update the Computer**: Update `NewComputer` and `newHasher` in `internal/hash/hash.go` to support the new algorithm.
3.  **Update Validation**: Add the new algorithm name to `ValidAlgorithms` in `internal/config/defaults.go`.
4.  **Smart Detection**: Update `detectHashAlgorithm` in `internal/config/cli.go` and `DetectHashAlgorithm` in `internal/hash/hash.go` with the expected hex string length of the new algorithm.
5.  **Tests**: Add a test case to `internal/hash/hash_test.go` and `internal/config/config_cli_test.go`.

## Adding a New CLI Flag

To add a new flag (e.g., `--follow-symlinks`):

1.  **Update the Config Struct**: Add a field to the `Config` struct in `internal/config/types.go`.
2.  **Define the Flag**: Add the flag definition to `defineFlags` in `internal/config/cli.go`.
3.  **Default Value**: If needed, set a default value in `DefaultConfig()` in `internal/config/defaults.go`.
4.  **Environmental Mapping**: Add a corresponding field to `EnvConfig` in `internal/config/types.go` and update `LoadEnvConfig` and `ApplyEnvConfig` in `internal/config/env.go`.
5.  **Config File Mapping**: Add a field to `ConfigFile` struct in `internal/config/file.go` and update `ApplyConfigFile`.
6.  **Implementation**: Use the new field in `cmd/chexum/main.go` or other relevant packages.
7.  **Documentation**: Update `docs/user/command-reference.md`.
