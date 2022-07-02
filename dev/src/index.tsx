import React from 'react'
import { render } from 'react-dom'

import { createOvermind } from 'overmind'
import { Provider } from 'overmind-react'
import { config } from './overmind'
import App from './App'
import { HashRouter } from 'react-router-dom'
import DevelopmentMode from './DevelopmentMode'

const overmind = createOvermind(config, {
    // Enable devtools by setting the below to ex. 'devtools: "localhost:3301"'
    // then run 'npx overmind-devtools@latest' to start the devtools
    devtools: "localhost:3301",
})


const DEVELOPMENT_MODE = (process.env.NODE_ENV === 'development' || process.env.NODE_ENV === 'test') && window.location.hostname === 'localhost'

render((
    <Provider value={overmind}>
        <HashRouter>
            {DEVELOPMENT_MODE && <DevelopmentMode />}
            <App />
        </HashRouter>
    </Provider>
), document.getElementById('root'))
