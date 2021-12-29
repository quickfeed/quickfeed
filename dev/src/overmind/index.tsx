import { createStateHook, createActionsHook, createEffectsHook } from 'overmind-react'
import { state } from './state'
import * as actions from './actions'
import * as effects from './effects'
import * as course from './namespaces/course'
import * as review from './namespaces/review'
import { IContext } from 'overmind'
import { merge, namespaced } from 'overmind/config';

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


export const useAppState = createStateHook<Context>()
export const useActions = createActionsHook<Context>()
export const useGrpc = createEffectsHook<Context>()
