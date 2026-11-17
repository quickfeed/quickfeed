import { useEffect, useRef } from "react"
import { useActions } from "../../overmind"
import type { Alert as AlertType } from "../../overmind/state"

/** Alert displays a countdown and dismisses the notification after its delay or when clicked. */
const Alert = ({ alert }: { alert: AlertType }) => {
    const circleRef = useRef<SVGCircleElement>(null)
    const actions = useActions().global
    const { id: alertID, delay } = alert

    useEffect(() => {
        let frame: number | undefined
        const circle = circleRef.current

        const id = setTimeout(() => {
            actions.popAlert({ id: alertID })
        }, delay)

        if (circle) {
            const circumference = circle.getTotalLength()

            circle.style.strokeDasharray = `${circumference}px`
            circle.style.strokeDashoffset = `${circumference}px`

            const start = Date.now()
            const animate = () => {
                const elapsed = Date.now() - start
                const strokeDashoffset = (elapsed / delay) * circumference
                circle.style.strokeDashoffset = `${strokeDashoffset}px`
                if (elapsed < delay) {
                    frame = requestAnimationFrame(animate)
                }
            }
            frame = requestAnimationFrame(animate)
        }

        return () => {
            clearTimeout(id)
            if (frame !== undefined) {
                cancelAnimationFrame(frame)
            }
        }
    }, [actions, alertID, delay])

    return (
        <div className={`alert alert-${alert.color}`} role="button" aria-hidden="true" style={{ marginTop: "20px", whiteSpace: "pre-wrap" }} onClick={() => actions.popAlert(alert)}>
            <svg viewBox="0 0 50 50" style={{ width: 20, height: 20, marginRight: 20 }}>
                <circle ref={circleRef} cx={25} cy={25} r={20} strokeWidth={5} fill="none" stroke="#000" />
            </svg>
            {alert.text}
        </div>
    )
}

export default Alert
