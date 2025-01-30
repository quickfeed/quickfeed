import React from 'react'
import { render } from 'react-dom'

import { createOvermind } from 'overmind'
import { Provider } from 'overmind-react'
import { config } from './overmind'
import App from './App'
import { BrowserRouter } from 'react-router-dom'
import './style.scss'

(BigInt.prototype as any).toJSON = function () { // skipcq: JS-0323
    return this.toString()
}

const overmind = createOvermind(config, {
    // Enable devtools by setting the below to ex. 'devtools: "localhost:3301"'
    // then run 'npx overmind-devtools@latest' to start the devtools
    devtools: "localhost:3301",
})

if (process.env.NODE_ENV === "development") {
    const eventSource = new EventSource("/watch")
    eventSource.onmessage = () => {
        setTimeout(() => {
            location.reload()
        }, 500)
    }
    // TODO: connection times out after 2 minutes, need to reconnect
    eventSource.onerror = () => console.error("could not connect to server-sent events")
}

render((
    <Provider value={overmind}>
        <BrowserRouter>
            <App />
        </BrowserRouter>
    </Provider>
), document.getElementById('root'))