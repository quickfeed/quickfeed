import { useParams } from "react-router"

/** getCourseID returns the course ID determined by the current route */
export const useCourseID = (): bigint => {
    const route = useParams<{ id?: string }>()
    try {
        return route.id ? BigInt(route.id) : BigInt(0)
    } catch (e) {
        return BigInt(0)
    }
}
