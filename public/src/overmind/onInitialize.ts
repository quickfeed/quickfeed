import { OnInitialize } from 'overmind'

export const onInitialize: OnInitialize = async ({
    state,
    actions,
    effects
  }, overmind) => {
    actions.setupUser().then(success => {
      console.log("print")
    })
  }