
import { IConfig } from 'overmind'
import {createActionsHook, createEffectsHook, createHook, createReactionHook, createStateHook} from 'overmind-react'
import { state } from './state'
import * as actions from './actions'
import * as effects from './effects'


export const config = {
    state,
    actions,
    effects
}

declare module 'overmind' {
    interface Config extends IConfig<typeof config> {}
}


export const useOvermind = createHook<typeof config>()
export const useState = createStateHook<typeof config>()
export const useActions = createActionsHook<typeof config>()
export const useEffects = createEffectsHook<typeof config>()
export const useReaction = createReactionHook<typeof config>()