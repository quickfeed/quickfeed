import React, { useEffect, useRef } from "react"
import { Alert as AlertType } from "../../overmind/state"
import { useActions } from "../../overmind"

/** Alert is a component that displays a single alert.
 *
 *  If the alert has a delay property, the alert will be removed after the delay.
 *  In this case, an animated circle will be displayed to indicate the time remaining before the alert is removed.
 *
 *  The alert is removed if the alert is clicked, or if the delay has passed.
 *
 *  @param alert - The alert to be displayed
 */
const Alert = ({ alert }: { alert: AlertType }) => {
    const circleRef = useRef<SVGCircleElement>(null)
    const actions = useActions().global

    useEffect(() => {
        let id: ReturnType<typeof setTimeout>
        if (alert.delay) {
            const circle = circleRef.current

            // Remove the alert after the delay
            id = setTimeout(() => {
                actions.popAlert(alert)
            }, alert.delay)

            if (circle) {
                const delay: number = alert.delay
                const circumference = circle.getTotalLength()

                circle.style.strokeDasharray = `${circumference}px`
                circle.style.strokeDashoffset = `${circumference}px`

                const start = Date.now()
                const animate = () => {
                    const elapsed = Date.now() - start
                    const strokeDashoffset = (elapsed / delay) * circumference
                    circle.style.strokeDashoffset = `${strokeDashoffset}px`
                    if (elapsed < delay) {
                        requestAnimationFrame(animate)
                    }
                }
                requestAnimationFrame(animate)
            }
        }

        return () => {
            if (id) {
                // If the alert is removed (by clicking the alert) before
                // the delay has passed, the timeout will be cleared.
                clearTimeout(id)
            }
        }
    }, [actions, alert])

    return (
        <div className={`alert alert-${alert.color}`} role="button" aria-hidden="true" style={{ marginTop: "20px", whiteSpace: "pre-wrap" }} onClick={() => actions.popAlert(alert)}>
            {alert.delay && (
                <svg viewBox="0 0 50 50" style={{ width: 20, height: 20, marginRight: 20 }}>
                    <circle ref={circleRef} cx={25} cy={25} r={20} strokeWidth={5} fill="none" stroke="#000" />
                </svg>
            )}
            {alert.text}
        </div>
    )
}

export default Alert
