import { Fragment, useEffect, useState } from "react"
import usePrefersReducedMotion from "../../hooks/usePrefersReducedMotion"
import { previewAssignment, previewAttempts, previewScoreLimit } from "./previewData"

type Step = {
    icon: string
    label: string
    /** FontAwesome animation played while the step is the active one. */
    motion: string
}

const steps: Step[] = [
    { icon: "fa-cloud-arrow-up", label: "Student pushes code", motion: "fa-bounce" },
    { icon: "fa-gears", label: "QuickFeed builds and tests", motion: "fa-spin" },
    // The feedback step shows the running score on a dial instead of an icon.
    { icon: "", label: "Student sees feedback", motion: "" },
    { icon: "fa-graduation-cap", label: "Teacher grades", motion: "" },
]

const scoreStep = 2

type Stage = {
    /** Index of the pipeline step that is active. */
    step: number
    /** Total score on the dial while the stage is shown. */
    score: number
    /** Points this push earned, shown floating above the feedback step. */
    gain: number
    approved: boolean
    /** Which push of previewAttempts the stage belongs to, counting from one. */
    attempt: number
    /** How long the stage is shown, in milliseconds. */
    hold: number
}

// The stages replay a student working towards a passing score: every push walks
// the pipeline and leaves a higher score behind, and the teacher only grades
// once a push clears previewScoreLimit. A push carries the previous score until
// its tests have run, so the dial advances when the feedback arrives.
const stages: Stage[] = previewAttempts
    .flatMap((score, index): Stage[] => {
        const previous = index === 0 ? 0 : previewAttempts[index - 1]
        const shared = { gain: score - previous, approved: false, attempt: index + 1 }
        return [
            { ...shared, step: 0, score: previous, hold: 1000 },
            { ...shared, step: 1, score: previous, hold: 1600 },
            { ...shared, step: 2, score, hold: 1900 },
        ]
    })
    .concat({
        step: 3,
        score: previewAttempts[previewAttempts.length - 1],
        gain: 0,
        approved: true,
        attempt: previewAttempts.length,
        hold: 3400,
    })

const finalStage = stages[stages.length - 1]

/**
 * WorkflowPreview walks the QuickFeed pipeline: a student pushes, the gears
 * turn, the score a push earned floats up and lands on the dial, and once the
 * total clears the assignment's score limit the teacher's approval mark
 * appears. Every color is a theme token, so the simulation follows the theme.
 */
const WorkflowPreview = () => {
    const prefersReducedMotion = usePrefersReducedMotion()
    const [index, setIndex] = useState(0)

    useEffect(() => {
        if (prefersReducedMotion) {
            return
        }
        const timer = setTimeout(() => setIndex(current => (current + 1) % stages.length), stages[index].hold)
        return () => clearTimeout(timer)
    }, [index, prefersReducedMotion])

    // Without animation the pipeline rests on its outcome rather than freezing
    // on the first push, which would read as a workflow that never finishes.
    const stage = prefersReducedMotion ? finalStage : stages[index]
    const animate = !prefersReducedMotion

    return (
        <div className="mt-12 space-y-5">
            <ol className="flex items-start">
                {steps.map((step, stepIndex) => (
                    <Fragment key={step.label}>
                        {stepIndex > 0 ? <Connector filled={stage.approved || stage.step >= stepIndex} approved={stage.approved} /> : null}
                        <li className="relative flex w-18 shrink-0 flex-col items-center gap-3 text-center sm:w-28 lg:w-36">
                            {stepIndex === scoreStep && animate && stage.step === scoreStep && stage.gain > 0 ? (
                                <FloatingGain key={`gain-${index}`} gain={stage.gain} />
                            ) : null}
                            {stepIndex === scoreStep ? (
                                <ScoreDial score={stage.score} state={stepState(stage, stepIndex)} />
                            ) : (
                                <IconBubble step={step} state={stepState(stage, stepIndex)} animate={animate} />
                            )}
                            <span className="text-xs font-medium leading-snug text-base-content/70 lg:text-sm">{step.label}</span>
                        </li>
                    </Fragment>
                ))}
            </ol>
            <p className="text-sm text-base-content/60" aria-hidden="true">
                {stage.approved
                    ? `${previewAssignment} approved after ${stage.attempt} pushes`
                    : `${previewAssignment} · push ${stage.attempt} of ${previewAttempts.length} · ${previewScoreLimit} % to pass`}
            </p>
        </div>
    )
}

type StepState = "idle" | "active" | "done" | "approved"

const stepState = (stage: Stage, stepIndex: number): StepState => {
    if (stage.approved) {
        return "approved"
    }
    if (stepIndex === stage.step) {
        return "active"
    }
    return stepIndex < stage.step ? "done" : "idle"
}

const bubbleClasses: Record<StepState, string> = {
    idle: "bg-base-300 text-base-content/40",
    done: "bg-primary/25 text-primary",
    active: "bg-primary text-primary-content scale-110 ring-4 ring-primary/25 shadow-lg",
    approved: "bg-success text-success-content ring-4 ring-success/25",
}

const IconBubble = ({ step, state, animate }: { step: Step, state: StepState, animate: boolean }) => (
    <span className={`relative flex size-14 items-center justify-center rounded-full transition-all duration-300 ${bubbleClasses[state]}`}>
        <i className={`fas ${step.icon} text-lg ${animate && state === "active" ? step.motion : ""}`} aria-hidden="true" />
        {state === "approved" && step.icon === "fa-graduation-cap" ? <ApprovalMark animate={animate} /> : null}
    </span>
)

// The dial is drawn on a circle of radius 15.9155, whose circumference is 100,
// so the score maps straight onto the dash pattern. The gap in the track marks
// the score limit the student has to reach.
const ScoreDial = ({ score, state }: { score: number, state: StepState }) => {
    const passed = score >= previewScoreLimit
    return (
        <span className={`relative flex size-14 items-center justify-center rounded-full bg-base-200 transition-all duration-300 ${state === "active" ? "scale-110 shadow-lg" : ""}`}>
            <svg viewBox="0 0 36 36" className="absolute inset-0 size-full -rotate-90" aria-hidden="true">
                <circle cx="18" cy="18" r="15.9155" fill="none" strokeWidth="3" className="stroke-base-300" />
                <circle
                    cx="18" cy="18" r="15.9155" fill="none" strokeWidth="3" strokeLinecap="round"
                    strokeDasharray="100" strokeDashoffset={100 - score}
                    className={`transition-[stroke-dashoffset] duration-700 ease-out ${passed ? "stroke-success" : "stroke-primary"}`}
                />
                <circle
                    cx="18" cy="18" r="15.9155" fill="none" strokeWidth="3"
                    strokeDasharray="1 99" strokeDashoffset={-previewScoreLimit}
                    className="stroke-base-content/40"
                />
            </svg>
            <span className={`tabular-nums text-sm font-bold ${passed ? "text-success" : "text-base-content/70"}`}>
                {score}%
            </span>
        </span>
    )
}

const FloatingGain = ({ gain }: { gain: number }) => (
    <span className="animate-float-up pointer-events-none absolute left-1/2 top-0 -translate-x-1/2 text-xl font-bold text-success">
        +{gain}
    </span>
)

const ApprovalMark = ({ animate }: { animate: boolean }) => (
    <span className={`absolute -right-1 -top-1 flex size-6 items-center justify-center rounded-full bg-success text-success-content ring-2 ring-base-100 ${animate ? "animate-pop-in" : ""}`}>
        <i className="fas fa-check text-xs" aria-hidden="true" />
    </span>
)

const Connector = ({ filled, approved }: { filled: boolean, approved: boolean }) => (
    <li aria-hidden="true" className="mt-7 h-1.5 min-w-2 flex-1 overflow-hidden rounded-full bg-base-300">
        <div className={`h-full rounded-full transition-all duration-500 ${filled ? "w-full" : "w-0"} ${approved ? "bg-success" : "bg-primary"}`} />
    </li>
)

export default WorkflowPreview
