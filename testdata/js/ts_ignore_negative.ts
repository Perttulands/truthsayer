// Negative: @ts-ignore with explanation or @ts-expect-error
// @ts-ignore legacy API returns untyped data
const x = legacyApi();

// @ts-expect-error testing invalid input
const y = processInput(null);
