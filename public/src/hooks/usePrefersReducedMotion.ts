import { useEffect, useState } from "react"

const query = "(prefers-reduced-motion: reduce)"

/** usePrefersReducedMotion reports whether the viewer has asked for reduced motion. */
const usePrefersReducedMotion = (): boolean => {
    const [prefersReducedMotion, setPrefersReducedMotion] = useState(
        () => window.matchMedia?.(query).matches ?? false
    )

    useEffect(() => {
        const mediaQuery = window.matchMedia?.(query)
        if (!mediaQuery?.addEventListener) {
            return
        }
        const update = (event: MediaQueryListEvent) => setPrefersReducedMotion(event.matches)
        mediaQuery.addEventListener("change", update)
        return () => mediaQuery.removeEventListener("change", update)
    }, [])

    return prefersReducedMotion
}

export default usePrefersReducedMotion
