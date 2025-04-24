import { createRoot } from "react-dom/client"
import React from 'react'
import { createOvermind } from 'overmind'
import { Provider } from 'overmind-react'
import { config } from './overmind'
import App from './App'
import { BrowserRouter } from 'react-router-dom'
import './style.scss'
import './tailwind.css'

(BigInt.prototype as any).toJSON = function () { // skipcq: JS-0323
    return this.toString()
}

const overmind = createOvermind(config, {
    // Enable devtools by setting the below to ex. 'devtools: "localhost:3031"'
    // then run 'npx overmind-devtools@latest' to start the devtools
    devtools: "localhost:3031",
})

if (process.env.NODE_ENV === "development") {

    // EventSource will automatically try to reconnect if the connection is lost
    const eventSource = new EventSource("/watch")
    eventSource.onmessage = () => {
        setTimeout(() => {
            location.reload()
        }, 200)
    }
    eventSource.onerror = () => console.error("could not connect to server-sent events")
}

const rootDocument = document.getElementById('root')
if (rootDocument) {
    const root = createRoot(rootDocument)

    root.render((<Provider value={overmind}>
        <BrowserRouter>
            <App />
        </BrowserRouter>
    </Provider>))
} else {
    throw new Error('Could not find root element with id "root"')
}
