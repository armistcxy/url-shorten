import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";
import { Box, CircularProgress, Typography } from "@mui/material";

const Preview: React.FC = () => {
  const { id } = useParams();
  const [originalUrl, setOriginalUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [timer, setTimer] = useState(10);

  useEffect(() => {
    if (id) {
      const fetchOriginalUrl = async () => {
        try {
          const response = await axios.get(`http://localhost/short/${id}`);
          setOriginalUrl(response.data.origin);
        } catch (err) {
          setError("Error fetching the original URL.");
        }
      };

      fetchOriginalUrl();
    }
  }, [id]);

  useEffect(() => {
    if (originalUrl) {
      const countdown = setInterval(() => {
        setTimer((prev) => prev - 1);
      }, 1000);

      if (timer === 0) {
        window.location.href = originalUrl;
      }

      return () => clearInterval(countdown);
    }
  }, [originalUrl, timer]);

  const handleLinkClick = () => {
    window.location.href = originalUrl!;
  };

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        height: "100vh",
        textAlign: "center",
        padding: 2,
      }}
    >
      <Typography variant="h4" sx={{ marginBottom: 2 }}>
        Redirecting...
      </Typography>
      {error ? (
        <Typography variant="h6" sx={{ color: "red", marginBottom: 2 }}>
          {error}
        </Typography>
      ) : (
        <>
          <Typography variant="h6" sx={{ marginBottom: 2 }}>
            You are being redirected to:
          </Typography>
          <a
            href={originalUrl!}
            target="_blank"
            rel="noopener noreferrer"
            onClick={handleLinkClick}
            style={{
              fontSize: "1.5rem",
              color: "blue",
              textDecoration: "none",
              cursor: "pointer",
              transition: "text-decoration 0.3s ease",
            }}
            onMouseEnter={(e) => (e.target.style.textDecoration = "underline")}
            onMouseLeave={(e) => (e.target.style.textDecoration = "none")}
          >
            {originalUrl}
          </a>
          <Typography variant="body1" sx={{ marginBottom: 2 }}>
            You will be redirected in {timer} seconds...
          </Typography>
          <CircularProgress sx={{ marginTop: 2 }} />
        </>
      )}
    </Box>
  );
};

export default Preview;
