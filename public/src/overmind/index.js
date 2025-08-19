import { createStateHook, createActionsHook, createEffectsHook } from 'overmind-react';
import { state } from './state';
import * as review from './namespaces/review';
import * as feedback from './namespaces/feedback';
import * as global from './namespaces/global';
import { merge, namespaced } from 'overmind/config';
export const config = merge({
    state
}, namespaced({
    review,
    feedback,
    global
}));
export const useAppState = createStateHook();
export const useActions = createActionsHook();
export const useGrpc = createEffectsHook();
