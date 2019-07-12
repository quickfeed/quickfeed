import { Assignment } from "./ag_pb";

// getDeadline returns the lab assignment's deadline as a string,
// or no deadline if the assignment has an undefined deadline field.
// This is a workaround method due to strict null checking in typescript.
export function getDeadline(lab: Assignment): string {
    const deadline = lab.getDeadline();
    if (deadline) {
        return deadline.toString();
    }
    return "no deadline";
}
