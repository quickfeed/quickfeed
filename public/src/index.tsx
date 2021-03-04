import React from "react";
import ReactDOM from "react-dom"
import { render } from 'react-dom'

import { createOvermind } from "overmind";
import { Provider } from 'overmind-react'
import { config, useOvermind } from "./overmind";
import App from './App'

const overmind = createOvermind(config, {
    devtools: true
})



render((<Provider value={overmind}>
            <App />
        </Provider>
        ), document.getElementById('root'));