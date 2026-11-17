import { act, cleanup, fireEvent, render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import Alerts from "../components/alerts/Alerts"
import { Color } from "../Helpers"
import { initializeOvermind } from "./TestHelpers"

beforeEach(() => {
    vi.useFakeTimers()
    // jsdom does not implement SVG geometry or animate the countdown.
    Object.defineProperty(SVGElement.prototype, "getTotalLength", { configurable: true, value: () => 2 * Math.PI * 20 })
    vi.spyOn(window, "requestAnimationFrame").mockReturnValue(1)
    vi.spyOn(window, "cancelAnimationFrame").mockImplementation(() => {})
})

afterEach(() => {
    cleanup()
    Reflect.deleteProperty(SVGElement.prototype, "getTotalLength")
    vi.useRealTimers()
    vi.restoreAllMocks()
})

const renderAlerts = () => {
    const overmind = initializeOvermind({})
    const view = render(<Provider value={overmind}><Alerts /></Provider>)
    return { ...view, ...overmind }
}

describe("alert dismissal", () => {
    it.each([undefined, 0, -100, NaN, Infinity])("dismisses alerts with delay %s after the default five seconds", async delay => {
        const { actions, state } = renderAlerts()
        await act(async () => actions.global.alert({ text: "Notification", color: Color.RED, delay }))
        expect(state.alerts[0].delay).toBe(5000)
        act(() => vi.advanceTimersByTime(4999))
        expect(screen.getByText("Notification")).toBeDefined()
        await act(async () => vi.advanceTimersByTime(1))
        expect(screen.queryByText("Notification")).toBeNull()
        expect(state.alerts).toHaveLength(0)
    })

    it("preserves an explicit ten-second delay", async () => {
        const { actions } = renderAlerts()
        await act(async () => actions.global.alert({ text: "Long notification", color: Color.YELLOW, delay: 10000 }))
        act(() => vi.advanceTimersByTime(9999))
        expect(screen.getByText("Long notification")).toBeDefined()
        await act(async () => vi.advanceTimersByTime(1))
        expect(screen.queryByText("Long notification")).toBeNull()
    })

    it("expires each alert independently without restarting existing timers", async () => {
        const { actions } = renderAlerts()
        await act(async () => actions.global.alert({ text: "First notification", color: Color.GREEN }))
        act(() => vi.advanceTimersByTime(2000))
        await act(async () => actions.global.alert({ text: "Second notification", color: Color.RED }))
        await act(async () => vi.advanceTimersByTime(3000))
        expect(screen.queryByText("First notification")).toBeNull()
        expect(screen.getByText("Second notification")).toBeDefined()
        await act(async () => vi.advanceTimersByTime(2000))
        expect(screen.queryByText("Second notification")).toBeNull()
    })

    it("allows early dismissal and cancels the timer and animation", async () => {
        const { actions, state } = renderAlerts()
        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout")
        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout")
        await act(async () => actions.global.alert({ text: "Dismiss me", color: Color.GREEN }))
        const timeout = setTimeoutSpy.mock.results[setTimeoutSpy.mock.calls.findIndex(call => call[1] === 5000)].value
        await act(async () => fireEvent.click(screen.getByText("Dismiss me")))
        expect(state.alerts).toHaveLength(0)
        expect(clearTimeoutSpy).toHaveBeenCalledWith(timeout)
        expect(cancelAnimationFrame).toHaveBeenCalledWith(1)
    })

    it("cancels the timer and animation on unmount", async () => {
        const { actions, unmount, state } = renderAlerts()
        const clearTimeoutSpy = vi.spyOn(globalThis, "clearTimeout")
        const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout")
        await act(async () => actions.global.alert({ text: "Notification", color: Color.GREEN }))
        const timeout = setTimeoutSpy.mock.results[setTimeoutSpy.mock.calls.findIndex(call => call[1] === 5000)].value
        unmount()
        expect(clearTimeoutSpy).toHaveBeenCalledWith(timeout)
        expect(cancelAnimationFrame).toHaveBeenCalledWith(1)
        act(() => vi.advanceTimersByTime(5000))
        expect(state.alerts).toHaveLength(1)
    })
})
