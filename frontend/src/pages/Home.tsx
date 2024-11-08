import "../App.css";
import Header from "../components/Header";
import Shorten from "../components/Shorten";
import Footer from "../components/Footer";

const Home = () => {
  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-4xl mx-auto">
        <Header />
        <Shorten />
        <Footer />
      </div>
    </div>
  );
};

export default Home;
