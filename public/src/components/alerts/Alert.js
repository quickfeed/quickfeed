import React, { useEffect, useRef } from "react";
import { useActions } from "../../overmind";
const Alert = ({ alert }) => {
    const circleRef = useRef(null);
    const actions = useActions().global;
    useEffect(() => {
        let id;
        if (alert.delay) {
            const circle = circleRef.current;
            id = setTimeout(() => {
                actions.popAlert(alert);
            }, alert.delay);
            if (circle) {
                const delay = alert.delay;
                const circumference = circle.getTotalLength();
                circle.style.strokeDasharray = `${circumference}px`;
                circle.style.strokeDashoffset = `${circumference}px`;
                const start = Date.now();
                const animate = () => {
                    const elapsed = Date.now() - start;
                    const strokeDashoffset = (elapsed / delay) * circumference;
                    circle.style.strokeDashoffset = `${strokeDashoffset}px`;
                    if (elapsed < delay) {
                        requestAnimationFrame(animate);
                    }
                };
                requestAnimationFrame(animate);
            }
        }
        return () => {
            if (id) {
                clearTimeout(id);
            }
        };
    }, [actions, alert]);
    return (React.createElement("div", { className: `alert alert-${alert.color}`, role: "button", "aria-hidden": "true", style: { marginTop: "20px", whiteSpace: "pre-wrap" }, onClick: () => actions.popAlert(alert) },
        alert.delay && (React.createElement("svg", { viewBox: "0 0 50 50", style: { width: 20, height: 20, marginRight: 20 } },
            React.createElement("circle", { ref: circleRef, cx: 25, cy: 25, r: 20, strokeWidth: 5, fill: "none", stroke: "#000" }))),
        alert.text));
};
export default Alert;
