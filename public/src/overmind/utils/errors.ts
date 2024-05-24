import { Code } from "@bufbuild/connect"
import { AnyMessage } from "@bufbuild/protobuf"
import { Response } from "../../client"

/** Prompt contains the messages to display to the user when prompting for confirmation. */
export namespace Prompt {
    export const GroupDelete = "Are you sure you want to delete this group?"
    export const GroupRepoNotEmpty = "Warning: The group repository is not empty. Do you still want to delete the group and its corresponding GitHub repository?"
    export const EnrollmentRepoNotEmpty = "Warning: The enrollment repository is not empty. Do you still want to delete the enrollment and enrollment repository?"
}

/** promptOnErrorResponse prompts the user with a warning if the response contains an error with the given code.
 *  If the user confirms the warning, the function returns null. Otherwise, it returns the error.
 *  The function is used to prompt the user before performing an action that may result in data loss.
 * @param response The response to check for errors.
 * @param errorCode The error code to check for.
 * @param message The message to display to the user.
 * @returns The error if the user did not confirm the warning, or null if the user did.
 *
*/
export function promptOnErrorResponse<T extends AnyMessage>(response: Response<T>, errorCode: Code, message: string) {
    if (response.error) {
        if (response.error.code === errorCode) {
            if (confirm(message)) {
                return null
            }
        }
    }
    return response.error
}
