// Negative: console.log in test file — expected
describe('handler', () => {
  it('should process data', () => {
    console.log('debugging test');
    expect(handler(input)).toBe(output);
  });
});
