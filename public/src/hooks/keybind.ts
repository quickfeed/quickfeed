import { useEffect } from 'react'

type KeyCombo = string | string[]
type ShortcutOptions = {
    keys: KeyCombo
    callback: (e: KeyboardEvent) => void
    enabled?: boolean
    preventDefault?: boolean
}

export function useKeyboardShortcut({
    keys,
    callback,
    enabled = true,
    preventDefault = true,
}: ShortcutOptions) {
    useEffect(() => {
        if (!enabled) return

        const keyList = Array.isArray(keys) ? keys : [keys]

        const handleKeyDown = (e: KeyboardEvent) => {
            const pressedKey = e.key.toLowerCase()
            if (keyList.includes(pressedKey)) {
                if (preventDefault) e.preventDefault()
                callback(e)
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => {
            window.removeEventListener('keydown', handleKeyDown)
        }
    }, [keys, callback, enabled, preventDefault])
}
