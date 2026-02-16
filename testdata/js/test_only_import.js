// Test fixture: test-only imports in production source
import { createMockUser } from '../__tests__/helpers';
import { stubApi } from './stubs';
import { render } from './test-utils';

export function getUser() {
  return createMockUser(); // should use real data
}
