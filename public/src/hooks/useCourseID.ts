import { useIDParam } from "./useIDParam"

/** useCourseID returns the course ID determined by the current route (0n if missing/invalid). */
export const useCourseID = (): bigint => useIDParam("id")
