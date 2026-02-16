// Positive fixture: deep optional chaining (>3 levels)

// Bad: 4 optional chain operators
const name = user?.profile?.address?.city?.name;

// Bad: 5 optional chain operators
const value = config?.settings?.database?.connection?.pool?.size;
