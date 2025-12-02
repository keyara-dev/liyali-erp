"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import Spinner from "@/components/ui/spinner";
import { createNewAccount } from "@/app/_actions/auth";
import {
  countries,
  formatCountryOption,
  formatCountrySelectDisplay,
  findCountryByDialCode,
} from "@/lib/countries";
import { PhoneInput } from "@/components/ui/phone-input";
import { Textarea } from "@/components/ui/textarea";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import Link from "next/link";
import { Input } from "@/components";

export default function Signup() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [shopName, setShopName] = useState("");
  const [description, setDescription] = useState("");
  const [whatsapp, setWhatsapp] = useState("");
  const [phoneInputValue, setPhoneInputValue] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState({
    password: false,
    confirmPassword: false,
  });
  const [message, setMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [referral, setReferral] = useState<string | null>(null);

  const [step, setStep] = useState(1); // 1: Store info, 2: Password creation
  const [showWhatsAppConfirmation, setShowWhatsAppConfirmation] =
    useState(false);

  // Capture referral from URL parameters
  useEffect(() => {
    const referralParam = searchParams.get("referral");
    if (referralParam) {
      setReferral(referralParam);
      console.log("Referral detected:", referralParam);
    }
  }, [searchParams]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage("");

    if (step === 1) {
      // First step: validate store info and show WhatsApp confirmation modal
      if (!email || !username || !shopName || !whatsapp) {
        setMessage(
          "Please fill in all required fields including WhatsApp number."
        );
        return;
      }
      setShowWhatsAppConfirmation(true);
      return;
    }

    if (step === 2) {
      // Second step: create account with password
      if (!password || !confirmPassword) {
        setMessage("Please enter and confirm your password.");
        return;
      }

      if (password !== confirmPassword) {
        setMessage("Passwords do not match. Please try again.");
        return;
      }

      if (password.length < 8) {
        setMessage("Password must be at least 8 characters long.");
        return;
      }

      setIsSubmitting(true);

      const response = await createNewAccount({
        email,
        username,
        shopName,
        description,
        whatsapp,
        password,
        referral,
      });

      if (response.success) {
        // Account created and user is automatically logged in
        // Redirect to dashboard
        router.push("/dashboard?account=new");
      } else {
        setMessage(`Error: ${response.message}`);
        setIsSubmitting(false);
      }

      setTimeout(() => {
        setIsSubmitting(false);
      }, 1000 * 60); // Simulate a delay of 1 minute
    }
  };

  const handleBack = () => {
    if (step > 1) {
      setStep(step - 1);
    }
    setMessage("");
  };

  const handleWhatsAppConfirm = () => {
    setShowWhatsAppConfirmation(false);
    setStep(2);
  };

  const handleWhatsAppEdit = () => {
    setShowWhatsAppConfirmation(false);
  };

  return (
    <div className="w-full max-w-md">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-light text-black mb-4">
          {step === 1 ? (
            <>
              Create your <span className="font-bold">store</span>
            </>
          ) : (
            <>
              Secure your <span className="font-bold">account</span>
            </>
          )}
        </h1>
        <p className="text-gray-600 leading-relaxed">
          {step === 1
            ? "Join thousands of creators who are already building their business with xclsv. Get started in less than 2 minutes."
            : "Create a strong password to protect your store and customer data."}
        </p>
      </div>

      {/* Referral Notification */}
      {referral && (
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center space-x-2">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-green-400"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div>
              <h3 className="text-sm font-medium text-green-800">
                You're invited by @{referral}!
              </h3>
              <p className="text-sm text-green-700">
                Complete your signup to earn 5 welcome points, and help @
                {referral} earn 10 referral points!
              </p>
            </div>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        {step === 1 ? (
          // Step 1: Store Information
          <>
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-black mb-2"
              >
                Email Address
              </label>
              <input
                type="email"
                id="email"
                className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
                placeholder="your@email.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            <div>
              <label
                htmlFor="username"
                className="block text-sm font-medium text-black mb-2"
              >
                Username{" "}
                <span className="text-xs text-gray-400">
                  (xclsv.com/@{username || "your-handle"})
                </span>
              </label>
              <input
                type="text"
                id="username"
                className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
                placeholder="your-handle"
                value={username}
                onChange={(e) =>
                  setUsername(
                    e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, "")
                  )
                }
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                • Only letters, numbers, - and _ allowed
              </p>
            </div>

            <div>
              <label
                htmlFor="shopName"
                className="block text-sm font-medium text-black mb-2"
              >
                Store Name
              </label>
              <input
                type="text"
                id="shopName"
                className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
                placeholder="Your Brand Name"
                value={shopName}
                onChange={(e) => setShopName(e.target.value)}
                required
              />
              <p className="text-xs text-gray-500 mt-1">
                This is your brand name and can contain any characters
              </p>
            </div>

            <div>
              <label
                htmlFor="description"
                className="block text-sm font-medium text-black mb-2"
              >
                Store Description{" "}
                <span className="text-gray-400">(Optional)</span>
              </label>
              <Textarea
                id="description"
                rows={4}
                // className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors resize-none placeholder:text-gray-400 text-black"
                placeholder="What do you sell? Tell your customers about your brand..."
                value={description}
                maxLength={250}
                showLimit={true}
                descriptionText="Appears on the store page."
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>

            <div>
              <label
                htmlFor="whatsapp"
                className="block text-sm font-medium text-black mb-2"
              >
                WhatsApp Number <span className="text-red-500">*</span>
              </label>
              <div className="flex gap-2">
                <Input
                  required
                  value={whatsapp}
                  onChange={(e) => {
                    setWhatsapp(e.target.value);
                    setPhoneInputValue(e.target.value);
                  }}
                  descriptionText="Required for store notifications and customer support"
                />
              </div>
            </div>

            <Button type="submit" disabled={isSubmitting} className="w-full">
              Continue to Password
            </Button>
          </>
        ) : (
          // Step 2: Password Creation
          <>
            <div className="bg-gray-50 p-4 rounded-lg mb-6">
              <h3 className="font-medium text-black mb-2">Store Information</h3>
              <div className="text-sm text-gray-600 space-y-1">
                <p>
                  <strong>Email:</strong> {email}
                </p>
                <p>
                  <strong>Username:</strong> @{username}
                </p>
                <p>
                  <strong>Store:</strong> {shopName}
                </p>
              </div>
            </div>

            <div className="relative">
              <label
                htmlFor="password"
                className="block text-sm font-medium text-black mb-2"
              >
                Password
              </label>
              <input
                type={showPassword.password ? "text" : "password"}
                id="password"
                className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
                placeholder="Enter a strong password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete={undefined}
              />
              <p className="text-xs text-gray-500 mt-1">
                Must be at least 8 characters long
              </p>
              {password && password.length > 0 && (
                <button
                  type="button"
                  className="absolute cursor-pointer right-3 top-[55%] transform -translate-y-1/2 text-gray-400"
                  onClick={() =>
                    setShowPassword((prev) => ({
                      ...prev,
                      password: !showPassword.password,
                    }))
                  }
                >
                  {showPassword.password ? (
                    <EyeOffIcon className="w-6 h-6 md:w-7 md:h-7" />
                  ) : (
                    <EyeIcon className="w-6 h-6 md:w-7 md:h-7" />
                  )}
                </button>
              )}
            </div>

            <div className="relative">
              <label
                htmlFor="confirmPassword"
                className="block text-sm font-medium text-black mb-2"
              >
                Confirm Password
              </label>
              <input
                type={showPassword.confirmPassword ? "text" : "password"}
                id="confirmPassword"
                className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent transition-colors placeholder:text-gray-400 text-black"
                placeholder="Confirm your password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
              {confirmPassword && confirmPassword.length > 0 && (
                <button
                  type="button"
                  className="absolute cursor-pointer right-3 top-[70%] transform -translate-y-1/2 text-gray-400"
                  onClick={() =>
                    setShowPassword((prev) => ({
                      ...prev,
                      confirmPassword: !showPassword.confirmPassword,
                    }))
                  }
                >
                  {showPassword.confirmPassword ? (
                    <EyeOffIcon className="w-6 h-6 md:w-7 md:h-7" />
                  ) : (
                    <EyeIcon className="w-6 h-6 md:w-7 md:h-7" />
                  )}
                </button>
              )}
            </div>

            <div className="flex gap-4">
              <Button
                type="button"
                variant="outline"
                onClick={handleBack}
                className="w-full bg-gray-100 text-black"
              >
                Back
              </Button>

              <Button
                type="submit"
                disabled={isSubmitting}
                isLoading={isSubmitting}
                className="w-full"
              >
                Register
              </Button>
            </div>
          </>
        )}
      </form>

      {message && (
        <div
          className={`mt-6 p-4 rounded-lg border ${
            message.includes("Error")
              ? "bg-red-50 border-red-200 text-red-700"
              : "bg-green-50 border-green-200 text-green-700"
          }`}
        >
          <p className="text-center text-sm">{message}</p>
        </div>
      )}

      <div className="mt-8 text-center">
        <p className="text-gray-600">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-black font-medium hover:underline"
          >
            Login
          </Link>
        </p>
      </div>

      {/* WhatsApp Confirmation Modal */}
      {showWhatsAppConfirmation && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md mx-auto shadow-2xl">
            <div className="text-center">
              <div className="text-4xl mb-4">📱</div>
              <h3 className="font-semibold text-gray-900 mb-3">
                Confirm your WhatsApp number
              </h3>
              <div className="bg-green-50 p-4 rounded-lg border border-green-200 mb-4">
                <p className="text-lg font-mono text-green-800">
                  {phoneInputValue}
                </p>
                <p className="text-sm text-gray-600 mt-1">
                  {findCountryByDialCode(phoneInputValue.split(" ")[0])?.name}
                </p>
              </div>
              <div className="text-sm text-gray-700 space-y-1 mb-6">
                <p>✓ Store order notifications</p>
                <p>✓ Account verification codes</p>
                <p>✓ Important store updates</p>
                <p>✓ Customer support access</p>
              </div>
            </div>

            <div className="flex gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={handleWhatsAppEdit}
                className="flex-1"
              >
                Edit Number
              </Button>
              <Button
                type="button"
                onClick={handleWhatsAppConfirm}
                className="flex-1"
              >
                Confirm & Continue
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
