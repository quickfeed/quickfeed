import { createStateHook, createActionsHook, createEffectsHook } from 'overmind-react'
import { state } from './state'
import * as actions from './actions'
import * as effects from './effects'
import * as review from './namespaces/review'
import { IContext } from 'overmind'
import { merge, namespaced } from 'overmind/config';

/* This is the main overmind configuration. */

/* To add a new namespace, add it to the namespaced object. */
export const config = merge(
    {
        state,
        actions,
        effects,
    },
    namespaced({
        review
    })
)


export type Context = IContext<{
    state: typeof config.state;
    actions: typeof config.actions;
    effects: typeof config.effects;
}>;

/* These are the overmind state hooks, which are used in components to access and modify the state. */
export const useAppState = createStateHook<Context>()
export const useActions = createActionsHook<Context>()
export const useGrpc = createEffectsHook<Context>()
