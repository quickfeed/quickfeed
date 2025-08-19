import { Code } from "@connectrpc/connect";
import { Prompt, promptOnErrorResponse } from "../../utils/errors";
export const isEmptyRepo = async ({ effects }, request) => {
    const response = await effects.global.api.client.isEmptyRepo(request);
    const prompt = request.groupID
        ? Prompt.GroupRepoNotEmpty
        : Prompt.EnrollmentRepoNotEmpty;
    return promptOnErrorResponse(response, Code.FailedPrecondition, prompt) === null;
};
