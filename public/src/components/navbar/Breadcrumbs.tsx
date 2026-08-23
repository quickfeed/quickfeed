import { Link, useLocation, useNavigate } from 'react-router'
import type { Assignment, Course } from "../../../proto/qf/types_pb"
import { ScreenSize } from "../../consts"
import useWindowSize from "../../hooks/windowsSize"
import { useActions, useAppState } from '../../overmind'


const Breadcrumbs = () => {
    const state = useAppState()
    const actions = useActions().global
    const location = useLocation()
    const navigate = useNavigate()
    const { width } = useWindowSize()
    const pathnames = location.pathname.split('/').filter(x => x)

    const handleDashboard = () => {
        actions.setActiveCourse(0n)
        navigate('/')
    }

    // Returns course name (or code if small screen)
    const resolveCourseName = (courses: Course[], courseId: string, width: number): string | null => {
        const course = courses.find(c => c.ID.toString() === courseId)
        if (!course) { return null }
        return width < ScreenSize.ExtraLarge ? course.code : course.name
    }

    // Returns assignment name (or null if not found)
    const resolveAssignmentName = (assignments: Assignment[], assignmentId: string): string | null => {
        const assignment = assignments.find(a => a.ID.toString() === assignmentId)
        return assignment?.name ?? null
    }

    // Both names follow from the route and the loaded courses, so they are
    // computed here rather than mirrored into state by an effect: pathnames is a
    // fresh array on every render, which made that effect run on every render.
    const [prefix, courseId, section, assignmentId] = pathnames
    const onCoursePath = prefix === 'course' && Boolean(courseId)
    const courseName = onCoursePath ? resolveCourseName(state.courses, courseId, width) : null
    const onLabPath = onCoursePath && (section === 'lab' || section === 'group-lab') && Boolean(assignmentId)
    const assignmentName = onLabPath ? resolveAssignmentName(state.assignments?.[courseId] ?? [], assignmentId) : null

    const segments: { label: string; to: string; last: boolean }[] = []
    pathnames.forEach((value, index) => {
        const to = `/${pathnames.slice(0, index + 1).join('/')}`
        // title case the path segment.
        let breadcrumbName = decodeURIComponent(value.charAt(0).toUpperCase() + value.slice(1))
        const last = index === pathnames.length - 1

        // skip the first path segment (e.g., 'course/ID').
        if (index === 0 && value === 'course') { return }
        // skip the second path segment (e.g., 'course/ID/lab/ID').
        if (index === 2 && (value === 'lab' || value === 'group-lab')) { return }
        // Replace 'course/ID' with 'course/Course Name' in the breadcrumb.
        if (index === 1 && courseName && pathnames[0] === 'course') { breadcrumbName = courseName }
        // Replace 'lab/ID' with 'lab/Assignment Name' in the breadcrumb.
        if (index === 3 && assignmentName && (pathnames[2] === 'lab' || pathnames[2] === 'group-lab')) { breadcrumbName = assignmentName }

        segments.push({ label: breadcrumbName, to, last })
    })

    return (
        <nav className="flex items-center font-mono text-md select-none">
            <span
                onClick={handleDashboard}
                className="cursor-pointer text-base-content/50 hover:text-primary transition-colors"
            >
                ~
            </span>
            {segments.map(({ label, to, last }) => (
                <span key={to} className="flex items-center">
                    <span className="mx-1 text-base-content/30">/</span>
                    {last ? (
                        <span className="text-primary font-semibold">{label}</span>
                    ) : (
                        <Link to={to} className="text-base-content/60 hover:text-primary transition-colors">
                            {label}
                        </Link>
                    )}
                </span>
            ))}
        </nav>
    )
}

export default Breadcrumbs
