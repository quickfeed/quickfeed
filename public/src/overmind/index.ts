import { createStateHook, createActionsHook, createEffectsHook } from 'overmind-react'
import { state } from './state'
// review, feedback and global does not expose themselves as an ES Module.
import * as review from './namespaces/review' // skipcq: JS-C1003
import * as feedback from './namespaces/feedback' // skipcq: JS-C1003
import * as global from './namespaces/global' // skipcq: JS-C1003
import { IContext } from 'overmind'
import { merge, namespaced } from 'overmind/config'

/* This is the main overmind configuration. */

/* To add a new namespace, add it to the namespaced object. */
export const config = merge(
    {
        state
    },
    namespaced({
        review,
        feedback,
        global
    })
)


export type Context = IContext<{
    state: typeof config.state
    actions: typeof config.actions
    effects: typeof config.effects
}>

/* These are the overmind state hooks, which are used in components to access and modify the state. */
export const useAppState = createStateHook<Context>()
export const useActions = createActionsHook<Context>()
export const useGrpc = createEffectsHook<Context>()
