import * as React from 'react'
import { render } from 'react-dom'

import { createOvermind } from 'overmind'
import { Provider } from 'overmind-react'
import { config } from './overmind'
import App from './App'
import { HashRouter } from 'react-router-dom'

const overmind = createOvermind(config, {
    devtools: false,
})



render((<Provider value={overmind}>
            <HashRouter>
                <App />
            </HashRouter>
        </Provider>
        ), document.getElementById('root'))
