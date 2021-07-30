import { createStateHook, createActionsHook } from 'overmind-react'
import { state } from './state'
import * as actions from './actions'
import * as effects from './effects'
import { IContext } from 'overmind'

export const config = {
    state,
    actions,
    effects,
}

export type Context = IContext<{
    state: typeof config.state;
    actions: typeof config.actions;
    effects: typeof config.effects;
}>;


export const useAppState = createStateHook<Context>()
export const useActions = createActionsHook<Context>()