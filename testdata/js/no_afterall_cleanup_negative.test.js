// Test fixture: proper cleanup with afterAll/afterEach

describe('Database tests', () => {
  beforeAll(() => {
    db = connectToDatabase();
  });

  afterAll(() => {
    db.close();
  });

  beforeEach(() => {
    resetData();
  });

  afterEach(() => {
    clearCache();
  });

  it('should query users', () => {
    expect(db.query('SELECT * FROM users')).toBeDefined();
  });
});
