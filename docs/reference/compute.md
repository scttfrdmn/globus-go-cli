# Compute Commands

Commands for distributed computing operations.

!!! info "Go CLI Exclusive"
    The Compute service is exclusively available in the Go CLI and not in the upstream Python CLI.

## globus compute endpoint list

List compute endpoints.

```bash
globus compute endpoint list
```

## globus compute endpoint show

Show details for a compute endpoint.

```bash
globus compute endpoint show ENDPOINT_ID
```

## globus compute function register

Register a function for execution.

```bash
globus compute function register [flags]
```

## globus compute function run

Execute a registered function.

```bash
globus compute function run FUNCTION_ID [flags]
```

## See Also

- [Command Reference](index.md)
