// Negative fixture: reasonable optional chaining

// Good: 2 levels
const city = user?.address?.city;

// Good: 3 levels (at threshold, not over)
const zip = user?.address?.city?.zip;

// Good: regular dot access (not optional)
const deep = a.b.c.d.e.f.g;

// Good: single optional chain
const name = user?.name;
