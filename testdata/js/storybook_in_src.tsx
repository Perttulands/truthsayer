// Positive: importing stories in production code
import { Default } from './Button.stories.tsx';
const stories = require('./Card.stories.js');

export function App() {
  return <Default />;
}
