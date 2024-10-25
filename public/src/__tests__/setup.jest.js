// For some reason the new proto types cause issues with Jest
// complaining about TextEncoder and TextDecoder not available
// on globalThis. This setup file fixes that issue for now.
const { TextDecoder, TextEncoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;