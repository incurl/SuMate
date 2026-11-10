package curated

// ProductStance is the position the product itself takes.
//
// The product's claim is that Su Dongpo can speak for a modern reader who
// cannot say what they feel.  That claim has a premise, and the premise is
// worth stating plainly -- because a tool that promises to find your voice
// has a duty to say when it cannot.
//
// The premise: the reader's situation is describable in the vocabulary the
// tree offers.  When it is not -- when what they feel has no name in any of
// the five dimensions -- the honest thing is to say so and to point at the
// nearest thing that does exist, rather than to pretend the fit is exact.
//
// This is the same discipline applied to the verses: a stance that does not
// hold must be declared, not quietly used.
type ProductStance struct {
	// Claim is what the product asserts about itself.
	Claim string `json:"claim"`
	// Premise is what must hold for the product's service to be genuine.
	Premise string `json:"premise"`
	// Shadow is what the product becomes when the premise fails.
	Shadow string `json:"shadow"`
	// Refusal is what the product does when it cannot serve.
	Refusal string `json:"refusal"`
	// Name is the display name of the stance.
	Name string `json:"name"`
}

// TheProductStance is the product's own position.
func TheProductStance() ProductStance {
	return ProductStance{
		Name: "东坡嘴替的位置",
		Claim: "东坡能替你说出那句你自己说不出的话——不是因为他的词更美，" +
			"而是因为他比你多活了几百年那些难。",
		Premise: "你此刻的处境，说得清；或者至少，说得出一半。" +
			"说不出的部分，往往正是最要紧的那部分，而东坡词恰好长于处理这种说不出的东西。",
		Shadow: "当工具开始替人说它没听懂的话，它就从嘴替变成了说教。" +
			"一句凑得很近但前提不成立的安慰，比一句诚实的「这里没有」更有害。",
		Refusal: "找不到真正贴合的那一句时，它会说明为什么找不到，" +
			"并给出最近的一句作为参照，而不是硬凑一个像样的答案。",
	}
}
