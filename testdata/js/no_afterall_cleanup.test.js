// Test fixture: beforeAll/beforeEach without matching cleanup

describe('Database tests', () => {
  beforeAll(() => {
    db = connectToDatabase();
  });

  beforeEach(() => {
    resetData();
  });

  it('should query users', () => {
    expect(db.query('SELECT * FROM users')).toBeDefined();
  });
});
