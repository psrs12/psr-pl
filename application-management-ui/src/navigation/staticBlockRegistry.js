import DeclinedScreen from './screens/DeclinedScreen.jsx';
import CancelledExpiredScreen from './screens/CancelledExpiredScreen.jsx';
import PostAcceptanceScreen from './screens/PostAcceptanceScreen.jsx';

export const STATIC_BLOCK_REGISTRY = {
  declined: DeclinedScreen,
  'cancelled-expired': CancelledExpiredScreen,
  'post-acceptance': PostAcceptanceScreen,
};
