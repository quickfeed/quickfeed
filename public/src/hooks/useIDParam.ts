import { useParams } from "react-router"
import { convertToBigInt } from "../Helpers"

/** useIDParam returns the named route parameter parsed as a bigint ID.
 *  Missing or malformed values (e.g. "7a") yield 0n, which downstream lookups treat as "not found". */
export const useIDParam = (name: string): bigint => {
    const params = useParams()
    return convertToBigInt(params[name])
}
