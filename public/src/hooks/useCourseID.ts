import { useParams } from "react-router"

/** getCourseID returns the course ID determined by the current route */
export const useCourseID = (): bigint => {
    const route = useParams<{ id?: string }>()
    return route.id ? BigInt(route.id) : BigInt(0)
}
