// Negative: importing stories in a story file — expected
import { Button } from './Button';

export default { title: 'Button', component: Button };
export const Default = () => <Button>Click</Button>;
