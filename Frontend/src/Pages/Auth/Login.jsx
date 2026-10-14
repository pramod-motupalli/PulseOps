import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import api from "../../API/api";

const Login = () => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState("");

    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setErrorMsg("");

        if (!email.trim()) {
            setErrorMsg("Please enter email");
            return;
        }

        if (!password) {
            setErrorMsg("Please enter password");
            return;
        }

        try {
            setLoading(true);

            const response = await api.post("/login", {
                email: email.trim(),
                password,
            });

            const token =
                response.data?.token ||
                response.data?.access_token;

            if (!token) {
                setErrorMsg("Login successful, but no authentication token was received.");
                return;
            }

            localStorage.setItem("token", token);

            if (response.data?.user) {
                localStorage.setItem(
                    "user",
                    JSON.stringify(response.data.user)
                );
            }

            navigate("/home");
        } catch (error) {
            console.error("Login error:", error);

            const msg =
                error.response?.data?.error ||
                error.response?.data?.message ||
                "Invalid credentials or server unavailable";

            setErrorMsg(msg);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-slate-200 max-w-[2160px] w-full min-h-screen mx-auto flex items-center justify-center p-4">
            <div className="w-full max-w-md bg-white p-8 sm:p-10 border border-gray-100 shadow-xl shadow-slate-200/60 rounded-2xl">

                {/* Header */}
                <div className="mb-8 text-center">
                    <h1 className="font-bold text-[#002244] text-3xl mb-2">
                        Welcome Back
                    </h1>

                    <p className="text-gray-500 text-sm">
                        Please enter your details to sign in.
                    </p>
                </div>

                {/* Error message */}
                {errorMsg && (
                    <div
                        role="alert"
                        className="mb-4 p-3.5 bg-red-50 border border-red-200 text-red-700 text-sm rounded-xl font-medium"
                    >
                        {errorMsg}
                    </div>
                )}

                {/* Login form */}
                <form
                    onSubmit={handleSubmit}
                    className="flex flex-col gap-5"
                >
                    {/* Email */}
                    <div className="flex flex-col gap-2">
                        <label
                            htmlFor="email"
                            className="text-sm font-semibold text-gray-700"
                        >
                            Email Address
                        </label>

                        <input
                            id="email"
                            name="email"
                            type="email"
                            value={email}
                            placeholder="you@example.com"
                            autoComplete="email"
                            required
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    {/* Password */}
                    <div className="flex flex-col gap-2">
                        <label
                            htmlFor="password"
                            className="text-sm font-semibold text-gray-700"
                        >
                            Password
                        </label>

                        <input
                            id="password"
                            name="password"
                            type="password"
                            value={password}
                            placeholder="Enter your password"
                            autoComplete="current-password"
                            required
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    {/* Submit button */}
                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full bg-[#002244] text-white font-semibold py-3 rounded-xl shadow-md shadow-blue-900/20 hover:bg-[#003366] hover:shadow-lg active:scale-[0.98] transition-all duration-200 disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                        {loading ? "Signing In..." : "Sign In"}
                    </button>
                </form>

                {/* Register link */}
                <p className="text-center text-sm text-gray-500 mt-8">
                    Don't have an account?{" "}
                    <Link
                        to="/register"
                        className="font-semibold text-[#002244] hover:text-blue-600 transition"
                    >
                        Create account
                    </Link>
                </p>
            </div>
        </div>
    );
};

export default Login;