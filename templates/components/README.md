# package components

## Best practices

- Ensure that "templ" components are **idempotent, pure function**.
  - For example, we could read global state directly inside a component, but we choose to actively get that state from within the component.  
