import { createOvermind } from 'overmind'
import { Provider } from 'overmind-react'
import { createRoot } from "react-dom/client"
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { config } from './overmind'
import './style.scss'

(BigInt.prototype as unknown as { toJSON: () => string }).toJSON = function () {
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
