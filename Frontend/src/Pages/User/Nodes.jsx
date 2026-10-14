import { useState, useEffect } from "react";
import SideBar from "../Layout/SideBar";
import Header from "../Layout/Header";
import api from "../../API/api";
import { Server, CheckCircle2, AlertCircle, Plus, X, Globe, Trash2, RefreshCw, Zap, ShieldAlert, Check } from "lucide-react";

const Nodes = () => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [loading, setLoading] = useState(true);
    const [submitting, setSubmitting] = useState(false);
    const [testing, setTesting] = useState(false);
    const [testResult, setTestResult] = useState(null);
    const [checkingNodeId, setCheckingNodeId] = useState(null);
    const [formData, setFormData] = useState({
        name: "",
        endpoint: "",
    });

    const [nodes, setNodes] = useState([]);

    const fetchNodes = async () => {
        try {
            setLoading(true);
            const response = await api.get("/nodes");
            if (Array.isArray(response.data)) {
                setNodes(response.data);
            }
        } catch (error) {
            console.error("Failed to fetch nodes:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefreshAll = async () => {
        try {
            setLoading(true);
            const response = await api.post("/nodes/check-all");
            if (Array.isArray(response.data)) {
                setNodes(response.data);
            }
        } catch (error) {
            await fetchNodes();
        } finally {
            setLoading(false);
        }
    };

    const handleCheckSingleNode = async (id) => {
        try {
            setCheckingNodeId(id);
            const response = await api.post(`/nodes/${id}/check`);
            if (response.data) {
                setNodes((prev) => prev.map((n) => (n.id === id ? response.data : n)));
            }
        } catch (error) {
            console.error("Failed to ping node:", error);
        } finally {
            setCheckingNodeId(null);
        }
    };

    const handleTestEndpoint = async () => {
        if (!formData.endpoint.trim()) {
            alert("Please enter an endpoint URL first.");
            return;
        }
        try {
            setTesting(true);
            setTestResult(null);
            const response = await api.post("/nodes/test", { endpoint: formData.endpoint.trim() });
            setTestResult(response.data);
        } catch (error) {
            setTestResult({
                status: "Critical",
                statusCode: 0,
                responseTimeMs: 0,
                errorMessage: error.response?.data?.error || "Connection failed",
            });
        } finally {
            setTesting(false);
        }
    };

    useEffect(() => {
        fetchNodes();
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            setSubmitting(true);
            const response = await api.post("/nodes", {
                name: formData.name.trim(),
                endpoint: formData.endpoint.trim(),
            });
            if (response.data) {
                setNodes((prev) => [response.data, ...prev]);
            }
            setFormData({ name: "", endpoint: "" });
            setTestResult(null);
            setIsModalOpen(false);
        } catch (error) {
            console.error("Failed to add node:", error);
            alert(error.response?.data?.error || "Failed to add endpoint");
        } finally {
            setSubmitting(false);
        }
    };

    const handleDeleteNode = async (id) => {
        if (!window.confirm("Are you sure you want to delete this endpoint?")) return;
        try {
            await api.delete(`/nodes/${id}`);
            setNodes((prev) => prev.filter((n) => n.id !== id && n.nodeId !== id));
        } catch (error) {
            console.error("Failed to delete node:", error);
            alert("Failed to delete endpoint");
        }
    };

    return (
        <div className="flex min-h-screen bg-slate-100 max-w-[2160px] mx-auto">
            <SideBar />
            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Header title="Services & Nodes" />

                <main className="flex-1 overflow-y-auto p-6 space-y-6">
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                        <div>
                            <h1 className="text-2xl font-bold text-gray-800 tracking-tight">Active Nodes & Clusters</h1>
                            <p className="text-sm text-gray-500">Monitor running services, verify API status codes, and analyze real-time response times.</p>
                        </div>

                        <button
                            onClick={() => {
                                setFormData({ name: "", endpoint: "" });
                                setTestResult(null);
                                setIsModalOpen(true);
                            }}
                            className="flex items-center gap-2 bg-[#002244] hover:bg-[#003366] text-white text-sm font-semibold px-4 py-2.5 rounded-xl transition shadow-sm w-fit active:scale-95"
                        >
                            <Plus className="w-4 h-4" /> Add Node
                        </button>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm flex items-center gap-4">
                            <div className="p-3 bg-blue-50 text-blue-600 rounded-xl">
                                <Server className="w-6 h-6" />
                            </div>
                            <div>
                                <p className="text-xs font-semibold text-gray-400">Total Nodes</p>
                                <p className="text-2xl font-bold text-gray-800">{nodes.length} Configured</p>
                            </div>
                        </div>

                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm flex items-center gap-4">
                            <div className="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
                                <CheckCircle2 className="w-6 h-6" />
                            </div>
                            <div>
                                <p className="text-xs font-semibold text-gray-400">Operational</p>
                                <p className="text-2xl font-bold text-gray-800">
                                    {nodes.filter((n) => n.status === "Healthy").length} Healthy
                                </p>
                            </div>
                        </div>

                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm flex items-center gap-4">
                            <div className="p-3 bg-amber-50 text-amber-600 rounded-xl">
                                <AlertCircle className="w-6 h-6" />
                            </div>
                            <div>
                                <p className="text-xs font-semibold text-gray-400">Warnings & Errors</p>
                                <p className="text-2xl font-bold text-gray-800">
                                    {nodes.filter((n) => n.status !== "Healthy").length} Issue{nodes.filter((n) => n.status !== "Healthy").length !== 1 ? "s" : ""}
                                </p>
                            </div>
                        </div>
                    </div>

                    <div className="bg-white border border-gray-200/80 rounded-2xl shadow-sm overflow-hidden">
                        <div className="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
                            <div>
                                <h3 className="font-bold text-gray-800">Cluster Node Inventory</h3>
                                <p className="text-xs text-gray-400 mt-0.5">Real-time status based on live API health checks</p>
                            </div>
                            <button
                                onClick={handleRefreshAll}
                                disabled={loading}
                                className="flex items-center gap-1.5 text-xs text-gray-600 hover:text-[#002244] font-medium transition px-3 py-1.5 rounded-lg border border-gray-200 hover:border-gray-300"
                            >
                                <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                                Re-check All
                            </button>
                        </div>

                        <div className="overflow-x-auto">
                            <table className="w-full text-left text-sm text-gray-600">
                                <thead className="bg-gray-50/75 text-xs font-semibold text-gray-400 border-b border-gray-100">
                                    <tr>
                                        <th className="px-6 py-3.5">Service Name</th>
                                        <th className="px-6 py-3.5">API Endpoint</th>
                                        <th className="px-6 py-3.5">Live Status</th>
                                        <th className="px-6 py-3.5">Status Code</th>
                                        <th className="px-6 py-3.5">Latency</th>
                                        <th className="px-6 py-3.5 text-right">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-100">
                                    {nodes.length === 0 && !loading && (
                                        <tr>
                                            <td colSpan="6" className="px-6 py-8 text-center text-gray-400">
                                                No endpoints configured yet. Click "Add Node" to start monitoring.
                                            </td>
                                        </tr>
                                    )}
                                    {nodes.map((node) => (
                                        <tr key={node.id} className="hover:bg-gray-50/60 transition">
                                            <td className="px-6 py-4 font-semibold text-gray-800 flex items-center gap-2.5">
                                                <Globe className="w-4 h-4 text-cyan-600 shrink-0" />
                                                <span>{node.name}</span>
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs text-gray-500 max-w-[260px] truncate" title={node.endpoint || node.url}>
                                                {node.endpoint || node.url}
                                            </td>
                                            <td className="px-6 py-4">
                                                <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold ${
                                                    node.status === "Healthy"
                                                        ? "bg-emerald-50 text-emerald-700 border border-emerald-200/60"
                                                        : node.status === "Warning"
                                                        ? "bg-amber-50 text-amber-700 border border-amber-200/60"
                                                        : "bg-red-50 text-red-700 border border-red-200/60"
                                                }`}>
                                                    <span className={`w-1.5 h-1.5 rounded-full ${
                                                        node.status === "Healthy"
                                                            ? "bg-emerald-500 animate-pulse"
                                                            : node.status === "Warning"
                                                            ? "bg-amber-500"
                                                            : "bg-red-500"
                                                    }`} />
                                                    {node.status}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs font-bold">
                                                <span className={`px-2 py-0.5 rounded ${
                                                    node.statusCode >= 200 && node.statusCode < 300
                                                        ? "bg-emerald-50 text-emerald-700 border border-emerald-200/60"
                                                        : node.statusCode >= 300 && node.statusCode < 400
                                                        ? "bg-blue-50 text-blue-700 border border-blue-200/60"
                                                        : node.statusCode >= 400 && node.statusCode < 500
                                                        ? "bg-amber-50 text-amber-700 border border-amber-200/60"
                                                        : "bg-red-50 text-red-700 border border-red-200/60"
                                                }`} title={node.errorMessage || ""}>
                                                    {node.statusCode > 0 ? node.statusCode : "ERR (Offline)"}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 font-mono text-xs text-gray-500">
                                                {node.responseTimeMs > 0 ? `${node.responseTimeMs} ms` : "-"}
                                            </td>
                                            <td className="px-6 py-4 text-right">
                                                <div className="flex items-center justify-end gap-1.5">
                                                    <button
                                                        onClick={() => handleCheckSingleNode(node.id)}
                                                        disabled={checkingNodeId === node.id}
                                                        className="p-1.5 text-gray-400 hover:text-[#002244] hover:bg-gray-100 rounded-lg transition"
                                                        title="Check Status Now"
                                                    >
                                                        <RefreshCw className={`w-4 h-4 ${checkingNodeId === node.id ? "animate-spin text-blue-600" : ""}`} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDeleteNode(node.id)}
                                                        className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
                                                        title="Delete Endpoint"
                                                    >
                                                        <Trash2 className="w-4 h-4" />
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </main>
            </div>

            {isModalOpen && (
                <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs flex items-center justify-center p-4">
                    <div className="bg-white w-full max-w-md rounded-2xl shadow-2xl border border-gray-100 overflow-hidden animate-in fade-in zoom-in-95 duration-150">
                        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                            <div>
                                <h3 className="font-bold text-gray-800 text-lg">Add New Endpoint</h3>
                                <p className="text-xs text-gray-400">Endpoint status will be verified on submission</p>
                            </div>
                            <button
                                onClick={() => setIsModalOpen(false)}
                                className="text-gray-400 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 transition"
                            >
                                <X className="w-5 h-5" />
                            </button>
                        </div>

                        <form onSubmit={handleSubmit} className="p-6 space-y-4">
                            <div>
                                <label className="block text-xs font-semibold text-gray-700 mb-1">
                                    Website / Service Name
                                </label>
                                <input
                                    type="text"
                                    required
                                    placeholder="e.g. Production API"
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    className="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-[#002244] focus:ring-2 focus:ring-blue-900/10 transition"
                                />
                            </div>

                            <div>
                                <div className="flex items-center justify-between mb-1">
                                    <label className="text-xs font-semibold text-gray-700">
                                        API Endpoint URL
                                    </label>
                                    <button
                                        type="button"
                                        onClick={handleTestEndpoint}
                                        disabled={testing || !formData.endpoint.trim()}
                                        className="text-xs font-semibold text-[#002244] hover:text-[#003366] flex items-center gap-1 disabled:opacity-50"
                                    >
                                        <Zap className={`w-3 h-3 ${testing ? "animate-pulse" : ""}`} />
                                        {testing ? "Testing..." : "Test Connection"}
                                    </button>
                                </div>
                                <input
                                    type="text"
                                    required
                                    placeholder="https://api.example.com/health"
                                    value={formData.endpoint}
                                    onChange={(e) => {
                                        setFormData({ ...formData, endpoint: e.target.value });
                                        setTestResult(null);
                                    }}
                                    className="w-full px-3.5 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:bg-white focus:border-[#002244] focus:ring-2 focus:ring-blue-900/10 transition font-mono text-xs"
                                />
                            </div>

                            {testResult && (
                                <div className={`p-3 rounded-xl border text-xs flex items-center justify-between ${
                                    testResult.status === "Healthy"
                                        ? "bg-emerald-50 border-emerald-200 text-emerald-800"
                                        : testResult.status === "Warning"
                                        ? "bg-amber-50 border-amber-200 text-amber-800"
                                        : "bg-red-50 border-red-200 text-red-800"
                                }`}>
                                    <div className="flex items-center gap-2">
                                        {testResult.status === "Healthy" ? (
                                            <Check className="w-4 h-4 text-emerald-600" />
                                        ) : (
                                            <ShieldAlert className="w-4 h-4 text-amber-600" />
                                        )}
                                        <span className="font-semibold">
                                            {testResult.statusCode > 0 ? `HTTP ${testResult.statusCode}` : "Connection Failed"}
                                            {testResult.status && ` • ${testResult.status}`}
                                        </span>
                                    </div>
                                    <span className="font-mono text-[11px]">
                                        {testResult.responseTimeMs > 0 ? `${testResult.responseTimeMs}ms` : ""}
                                    </span>
                                </div>
                            )}

                            <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-100">
                                <button
                                    type="button"
                                    onClick={() => setIsModalOpen(false)}
                                    className="px-4 py-2 text-sm font-semibold text-gray-600 hover:bg-gray-100 rounded-xl transition"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={submitting}
                                    className="px-5 py-2 text-sm font-semibold bg-[#002244] hover:bg-[#003366] text-white rounded-xl shadow-sm transition active:scale-95 disabled:opacity-75 flex items-center gap-2"
                                >
                                    {submitting && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
                                    {submitting ? "Verifying & Adding..." : "Add Endpoint"}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Nodes;
