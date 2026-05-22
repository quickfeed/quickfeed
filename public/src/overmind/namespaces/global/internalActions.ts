import { Code } from "@connectrpc/connect"
import type { Context } from "../.."
import type { RepositoryRequest } from "../../../../proto/qf/requests_pb"
import { Prompt, promptOnErrorResponse } from "../../utils/errors"

export const isEmptyRepo = async (
  { effects }: Context,
  request: RepositoryRequest
) => {
  const response = await effects.global.api.client.isEmptyRepo(request)
  const prompt = request.groupID
    ? Prompt.GroupRepoNotEmpty
    : Prompt.EnrollmentRepoNotEmpty

  return promptOnErrorResponse(response, Code.FailedPrecondition, prompt) === null
}
