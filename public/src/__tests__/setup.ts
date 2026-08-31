/// <reference types="node" />
import { TextDecoder, TextEncoder } from "node:util"

globalThis.TextEncoder ??= TextEncoder as typeof globalThis.TextEncoder
globalThis.TextDecoder ??= TextDecoder as typeof globalThis.TextDecoder
