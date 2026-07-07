import { useIDParam } from "./useIDParam"

/** useEnrollmentID returns the enrollment ID from the current route (0n if missing/invalid). */
export const useEnrollmentID = (): bigint => useIDParam("enrollmentID")
