import DeclinedScreen from './screens/DeclinedScreen.jsx';
import CancelledExpiredScreen from './screens/CancelledExpiredScreen.jsx';
import PostAcceptanceScreen from './screens/PostAcceptanceScreen.jsx';
import ErrorScreen from './screens/ErrorScreen.jsx';

export const STATIC_BLOCK_REGISTRY = {
  declined: DeclinedScreen,
  'cancelled-expired': CancelledExpiredScreen,
  'post-acceptance': PostAcceptanceScreen,
  error: ErrorScreen,
};
