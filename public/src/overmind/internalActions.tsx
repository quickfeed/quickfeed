import { Code } from "@bufbuild/connect"
import { Context } from "."
import { RepositoryRequest } from "../../proto/qf/requests_pb"
import { Prompt, promptOnErrorResponse } from "./utils/errors"

export const isEmptyRepo = async (
  { effects }: Context,
  request: Partial<RepositoryRequest>
) => {
  const response = await effects.api.client.isEmptyRepo(request)
  const prompt = request.groupID
    ? Prompt.GroupRepoNotEmpty
    : Prompt.EnrollmentRepoNotEmpty
  return promptOnErrorResponse(response, Code.FailedPrecondition, prompt) === null
}
