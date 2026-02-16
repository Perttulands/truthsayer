// Positive: Test library imports in non-test source file
import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';

export function setupComponent(Component) {
  return render(Component);
}
