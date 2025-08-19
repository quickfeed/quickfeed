export var Prompt;
(function (Prompt) {
    Prompt.GroupDelete = "Are you sure you want to delete this group?";
    Prompt.GroupRepoNotEmpty = "Warning: The group repository is not empty. Do you still want to delete the group and its corresponding GitHub repository?";
    Prompt.EnrollmentRepoNotEmpty = "Warning: The enrollment repository is not empty. Do you still want to delete the enrollment and enrollment repository?";
})(Prompt || (Prompt = {}));
export function promptOnErrorResponse(response, errorCode, message) {
    if (response.error) {
        if (response.error.code === errorCode) {
            if (confirm(message)) {
                return null;
            }
        }
    }
    return response.error;
}
