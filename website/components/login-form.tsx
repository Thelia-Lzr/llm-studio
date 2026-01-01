import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Welcome back</CardTitle>
          <CardDescription>
            Login with your GitHub account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form>
            <FieldGroup>
              <Field>
                <Button variant="outline" asChild>
                  <a href="/api/auth/github/login?return_to=/dashboard">
                    <svg viewBox="0 0 98 96" xmlns="http://www.w3.org/2000/svg">
                      <path
                        fill="currentColor"
                        fillRule="evenodd"
                        clipRule="evenodd"
                        d="M49 0C21.9 0 0 22.4 0 50c0 22.1 14.3 40.8 34.2 47.4 2.5.5 3.4-1.1 3.4-2.4 0-1.2 0-4.4-.1-8.6-13.9 3.1-16.8-6.9-16.8-6.9-2.3-5.9-5.6-7.5-5.6-7.5-4.6-3.2.3-3.1.3-3.1 5.1.4 7.8 5.3 7.8 5.3 4.5 7.8 11.8 5.5 14.7 4.2.5-3.3 1.8-5.5 3.2-6.8-11.1-1.3-22.8-5.7-22.8-25.2 0-5.6 2-10.2 5.2-13.8-.5-1.3-2.2-6.5.5-13.6 0 0 4.2-1.4 13.8 5.2 4-1.1 8.3-1.7 12.6-1.7 4.3 0 8.6.6 12.6 1.7 9.6-6.6 13.8-5.2 13.8-5.2 2.7 7.1 1 12.3.5 13.6 3.2 3.6 5.2 8.2 5.2 13.8 0 19.6-11.7 23.9-22.9 25.1 1.8 1.6 3.4 4.8 3.4 9.7 0 7-.1 12.6-.1 14.3 0 1.3.9 2.9 3.4 2.4C83.7 90.8 98 72.1 98 50 98 22.4 76.1 0 49 0Z"
                      />
                    </svg>
                    Continue with GitHub
                  </a>
                </Button>
              </Field>
              <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
                Or continue with
              </FieldSeparator>
              <Field>
                <FieldLabel htmlFor="email">Email</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  placeholder="m@example.com"
                  required
                />
              </Field>
              <Field>
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                  <a
                    href="#"
                    className="ml-auto text-sm underline-offset-4 hover:underline"
                  >
                    Forgot your password?
                  </a>
                </div>
                <Input id="password" type="password" required />
              </Field>
              <Field>
                <Button type="submit">Login</Button>
                <FieldDescription className="text-center">
                  Don&apos;t have an account? <a href="#">Sign up</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
      <FieldDescription className="px-6 text-center">
        By clicking continue, you agree to our <a href="#">Terms of Service</a>{" "}
        and <a href="#">Privacy Policy</a>.
      </FieldDescription>
    </div>
  )
}
