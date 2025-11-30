import SignupDisabled from "../_components/signup-disabled";
import { checkSignupAvailability } from "@/app/_actions/auth";
import Signup from "../_components/signup";

export default async function SignupPage() {
  const check = await checkSignupAvailability();

  if (!check?.data?.enabled) return <SignupDisabled />;

  return (
    <>
      <Signup />
    </>
  );
}
