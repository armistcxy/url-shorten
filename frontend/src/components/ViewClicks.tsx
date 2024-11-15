import React, { useState, useEffect } from "react";
import axios from "axios";
import { Box, Typography, CircularProgress } from "@mui/material";
import { Eye } from "lucide-react";

const API_BASE_URL = "http://localhost";
const POLLING_INTERVAL = 5000;

interface ViewClicksProps {
  urlId: string;
}

const ViewClicks: React.FC<ViewClicksProps> = ({ urlId }) => {
  const [clicks, setClicks] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchClicks = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/view/${urlId}`);
      setClicks(response.data.count);
      setError(null);
    } catch (err) {
      setError("Failed to fetch click count");
      console.error("Error fetching clicks:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClicks(); // Initial fetch

    // Set up polling
    const intervalId = setInterval(fetchClicks, POLLING_INTERVAL);

    // Cleanup on unmount
    return () => clearInterval(intervalId);
  }, [urlId]);

  if (loading) {
    return (
      <Box sx={{ display: "flex", alignItems: "center", p: 1 }}>
        <CircularProgress size={16} sx={{ color: "rgb(99 102 241)" }} />
      </Box>
    );
  }

  if (error || clicks === null) {
    return null;
  }

  return (
    <Box
      sx={{
        display: "flex",
        alignItems: "center",
        gap: 1,
        p: 1,
        borderRadius: "9999px",
        "&:hover": {
          backgroundColor: "rgba(99, 102, 241, 0.1)",
        },
      }}
    >
      <Eye className="w-4 h-4 text-indigo-500" />
      <Typography
        variant="body2"
        sx={{
          color: "rgb(75 85 99)",
          fontSize: "0.875rem",
        }}
      >
        {clicks} {clicks === 1 ? "click" : "clicks"}
      </Typography>
    </Box>
  );
};

export default ViewClicks;
