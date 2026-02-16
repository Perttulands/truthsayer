// Positive: jest mocks in source file
const mockFn = jest.fn(() => 42);
jest.mock('./database');
const spy = jest.spyOn(service, 'create');
