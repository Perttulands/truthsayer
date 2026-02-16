// Negative: jest mocks in test file — expected
jest.mock('./database');
const mockFn = jest.fn();
const spy = jest.spyOn(service, 'create');
