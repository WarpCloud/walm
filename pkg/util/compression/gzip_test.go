package compression

import (
	"k8s.io/klog"
	"testing"
)

func TestGzip(t *testing.T) {
	plain := `
Makaveli in this... Killuminati, all through your body
The blow's like a twelve gauge shotty
Uh, feel me!
And God said he should send his one begotten son
To lead the wild into the ways of the man
Follow me; eat my flesh, flesh and my flesh
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
I ain't a killer but don't push me
Revenge is like the sweetest joy next to getting pussy
Picture paragraphs unloaded, wise words being quoted
Peeped the weakness in the rap game and sewed it
Bow down, pray to God hoping that he's listening
Seeing niggas coming for me, to my diamonds, when they glistening
Now pay attention, rest in peace father
I'm a ghost in these killing fields
Hail Mary catch me if I go, let's go deep inside
The solitary mind of a madman who screams in the dark
Evil lurks, enemies, see me flee
Activate my hate, let it break, to the flame
Set trip, empty out my clip, never stop to aim
Some say the game is all corrupted, fucked in this shit
Stuck, niggas is lucky if we bust out this shit, plus
Mama told me never stop until I bust a nut
Fuck the world if they can't adjust
It's just as well, Hail Mary
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
Penitentiaries is packed with promise makers
Never realize the precious time the bitch niggas is wasting
Institutionalized I lived my life a product made to crumble
But too hardened for a smile, we're too crazy to be humble, we balling
Catch me father please, cause I'm falling, in the liquor store
That's the Hennessee I hear ya calling, can I get some more?
Hail 'til I reach Hell, I ain't scared
Mama checking in my bedroom; I ain't there
I got a head with no screws in it, what can I do
One life to live but I got nothing to lose, just me and you
On a one way trip to prison, selling drugs
We all wrapped up in this living, life as Thugs
To my homeboys in Clinton Max, doing they bid
Raise hell to this real shit, and feel this
When they turn out the lights, I'll be down in the dark
Thuggin eternal through my heart, now Hail Mary nigga
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
They got a APB, out on my Thug family
Since the Outlawz run these streets, like these skanless freaks
Our enemies die now, walk around half dead
Head down, K blasted off Hennessee and Thai
Trying it, mixed it, now I'm twisted blisted and high
Visions of me, Thug living getting me by
Forever live, and I multiply survived by Thugs
When I die they won't cry unless they coming with slugs
Peep the whole scene and whatever's going on around me
Brain kinda cloudy, smoked out feeling rowdy
Ready to wet the party up, and whoever in that motherfucker
Nasty new street, slugger my heat seeks suckers
On the regular mashing in a stolen black Ac Integ-ra
Cock back, sixty seconds 'til the draw that's when I'm dead in ya
Feet first, you got a nice gat but my heat's worse
From a Thug to preaching church, I gave you love now you eating dirt
Needing work, and I ain't the nigga to put you on
Cause word is bond when I was broke I had to hustle 'til dawn
That's when sun came up, there's only one way up
Hold ya head and stay up, to all my niggas get ya pay and weight up
If it's on then it's on, we break beat-breaks
Outlawz on a paper chase, can you relate
To this shit I don't got, be the shit I gotta take
Dealing with fate, hoping God don't close the gate
If it's on then it's on, we break beat-breaks
Outlawz on a paper chase, can you relate
To this shit I don't got, be the shit I gotta take
Dealing with fate, hoping God don't close the gate
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
We've been traveling on this wayward road
Long time 'til I be take a 'eavy load
But we ride, ride it like a bullet
Hail Mary, Hail Mary
We won't worry everything will come real
Free like the bird in the tree
We won't worry everything will come real
Yes we free like the bird in the tree
We running from the penitentiary
This is the time for we liberty
Hail Mary, Hail Mary
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
Come with me, Hail Mary
Run quick see, what do we have here
Now, do you wanna ride or die
La dadada, la la la la
Westside, Outlawz, Makaveli the Don, Solo, Killuminati, The 7 Days`
	compressed, err := GzipCompress(plain)
	if err != nil {
		t.Fail()
	}
	plain2, err := GzipDecompress(compressed)
	if plain != plain2 {
		t.Fail()
	}
	klog.Infof("compress ratio %f", float32(len(plain)-len(compressed))/float32(len(plain)))
}
