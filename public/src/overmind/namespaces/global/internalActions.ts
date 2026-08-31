import { Code } from "@connectrpc/connect"
import type { Context } from "../.."
import type { RepositoryRequest } from "../../../../proto/qf/requests_pb"
import { Prompt, promptOnErrorResponse } from "../../utils/errors"

export const isEmptyRepo = async (
  { effects }: Context,
  request: RepositoryRequest
) => {
  const response = await effects.global.api.client.isEmptyRepo(request)
  // The rawMessage contains the server's reason without the error code prefix,
  // e.g. "repository qf101-meling is 3 commits ahead".
  const reason = response.error?.rawMessage || "The repository is not empty"
  const prompt = request.groupID
    ? Prompt.GroupRepoNotEmpty(reason)
    : Prompt.EnrollmentRepoNotEmpty(reason)

  return promptOnErrorResponse(response, Code.FailedPrecondition, prompt) === null
}
