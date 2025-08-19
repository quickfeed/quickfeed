import { createRoot } from "react-dom/client";
import React from 'react';
import { createOvermind } from 'overmind';
import { Provider } from 'overmind-react';
import { config } from './overmind';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import './style.scss';
BigInt.prototype.toJSON = function () {
    return this.toString();
};
const overmind = createOvermind(config, {
    devtools: "localhost:3031",
});
if (process.env.NODE_ENV === "development") {
    const eventSource = new EventSource("/watch");
    eventSource.onmessage = () => {
        setTimeout(() => {
            location.reload();
        }, 200);
    };
    eventSource.onerror = () => console.error("could not connect to server-sent events");
}
const rootDocument = document.getElementById('root');
if (rootDocument) {
    const root = createRoot(rootDocument);
    root.render((React.createElement(Provider, { value: overmind },
        React.createElement(BrowserRouter, null,
            React.createElement(App, null)))));
}
else {
    throw new Error('Could not find root element with id "root"');
}
