import { STATIC_BLOCK_REGISTRY } from '../staticBlockRegistry.js';

export default function StaticBlockScreen({ descriptor, ctx }) {
  const Component = STATIC_BLOCK_REGISTRY[descriptor.block];
  return <Component descriptor={descriptor} ctx={ctx} />;
}
