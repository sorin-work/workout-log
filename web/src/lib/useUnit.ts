import { useAuth } from '../auth/context';
import type { Unit } from './units';

/** The signed-in user's preferred weight unit. */
export function useUnit(): Unit {
  return useAuth().user?.unit ?? 'kg';
}
