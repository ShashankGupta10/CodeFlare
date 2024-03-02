import type { Metadata } from "next";
import { Poppins } from "next/font/google";
import "./globals.css";

const poppins = Poppins({ subsets: ["latin"], weight: ["100", "200", "300", "400", "500", "600", "700", "800"] });

export const metadata: Metadata = {
  title: "CodeFlare | Your React Web App Deployment Platform",
  description: "CodeFlare is a cutting-edge web deployment platform tailored for modern React applications. Streamline your deployment process, manage versions, and optimize performance effortlessly. With intuitive features and seamless integration, CodeFlare empowers developers to deploy with confidence and efficiency.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={poppins.className}>{children}</body>
    </html>
  );
}
