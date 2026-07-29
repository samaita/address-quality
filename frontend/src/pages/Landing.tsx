import Hero from "@/components/landing/Hero"
import TrustStrip from "@/components/landing/TrustStrip"
import ExampleSection from "@/components/landing/ExampleSection"
import HowItWorks from "@/components/landing/HowItWorks"
import Features from "@/components/landing/Features"
import Limitations from "@/components/landing/Limitations"
import ClosingCTA from "@/components/landing/ClosingCTA"

export default function Landing() {
  return (
    <>
      <Hero />
      <TrustStrip />
      <ExampleSection />
      <HowItWorks />
      <Features />
      <Limitations />
      <ClosingCTA />
    </>
  )
}
