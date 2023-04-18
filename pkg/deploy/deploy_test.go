package deploy

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const expected = `
Fine. [to the liberal panelist] Sharon, the NEA is a loser. Yeah, it accounts for a penny
out of our paychecks, but he [gesturing to the conservative panelist] gets to hit you with it
anytime he wants. It doesn't cost money, it costs votes. It costs airtime and column inches.
You know why people don't like liberals? Because they lose. If liberals are so friggin’ smart,
how come they lose so GODDAM ALWAYS!
And [to the conservative panelist] with a straight face, you're going to tell students that
America's so starspangled awesome that we're the only ones in the world who have freedom?
Canada has freedom, Japan has freedom, the UK, France, Italy, Germany, Spain, Australia,
Belgium has freedom. Two hundred seven sovereign states in the world, like 180 of them
have freedom.
And you—sorority girl—yeah—just in case you accidentally wander into a voting booth one
day, there are some things you should know, and one of them is that there is absolutely no
evidence to support the statement that we're the greatest country in the world. We're seventh
in literacy, twenty-seventh in math, twenty-second in science, forty-ninth in life expectancy,
178th in infant mortality, third in median household income, number four in labor force, and
number four in exports. We lead the world in only three categories: number of incarcerated
citizens per capita, number of adults who believe angels are real, and defense spending,
where we spend more than the next twenty-six countries combined, twenty-five of whom are
allies. None of this is the fault of a 20-year-old college student, but you, nonetheless, are
without a doubt, a member of the WORST-period-GENERATION-period-EVER-period, so
when you ask what makes us the greatest country in the world, I don't know what the hell
you're talking about?! Yosemite?!!!
We sure used to be. We stood up for what was right! We fought for moral reasons, we passed
and struck down laws for moral reasons. We waged wars on poverty, not poor people. We
sacrificed, we cared about our neighbors, we put our money where our mouths were, and we
never beat our chest. We built great big things, made ungodly technological advances,
explored the universe, cured diseases, and cultivated the world's greatest artists and the
world's greatest economy. We reached for the stars, and we acted like men. We aspired to
intelligence; we didn't belittle it; it didn't make us feel inferior. We didn't identify ourselves by
who we voted for in the last election, and we didn't scare so easy. And we were able to be all
these things and do all these things because we were informed. By great men, men who were
revered. The first step in solving any problem is recognizing there is one—America is not the
greatest country in the world anymore.

Fine. [to the liberal panelist] Sharon, the NEA is a loser. Yeah, it accounts for a penny
out of our paychecks, but he [gesturing to the conservative panelist] gets to hit you with it
anytime he wants. It doesn't cost money, it costs votes. It costs airtime and column inches.
You know why people don't like liberals? Because they lose. If liberals are so friggin’ smart,
how come they lose so GODDAM ALWAYS!
And [to the conservative panelist] with a straight face, you're going to tell students that
America's so starspangled awesome that we're the only ones in the world who have freedom?
Canada has freedom, Japan has freedom, the UK, France, Italy, Germany, Spain, Australia,
Belgium has freedom. Two hundred seven sovereign states in the world, like 180 of them
have freedom.
And you—sorority girl—yeah—just in case you accidentally wander into a voting booth one
day, there are some things you should know, and one of them is that there is absolutely no
evidence to support the statement that we're the greatest country in the world. We're seventh
in literacy, twenty-seventh in math, twenty-second in science, forty-ninth in life expectancy,
178th in infant mortality, third in median household income, number four in labor force, and
number four in exports. We lead the world in only three categories: number of incarcerated
citizens per capita, number of adults who believe angels are real, and defense spending,
where we spend more than the next twenty-six countries combined, twenty-five of whom are
allies. None of this is the fault of a 20-year-old college student, but you, nonetheless, are
without a doubt, a member of the WORST-period-GENERATION-period-EVER-period, so
when you ask what makes us the greatest country in the world, I don't know what the hell
you're talking about?! Yosemite?!!!
We sure used to be. We stood up for what was right! We fought for moral reasons, we passed
and struck down laws for moral reasons. We waged wars on poverty, not poor people. We
sacrificed, we cared about our neighbors, we put our money where our mouths were, and we
never beat our chest. We built great big things, made ungodly technological advances,
explored the universe, cured diseases, and cultivated the world's greatest artists and the
world's greatest economy. We reached for the stars, and we acted like men. We aspired to
intelligence; we didn't belittle it; it didn't make us feel inferior. We didn't identify ourselves by
who we voted for in the last election, and we didn't scare so easy. And we were able to be all
these things and do all these things because we were informed. By great men, men who were
revered. The first step in solving any problem is recognizing there is one—America is not the
greatest country in the world anymore.

Fine. [to the liberal panelist] Sharon, the NEA is a loser. Yeah, it accounts for a penny
out of our paychecks, but he [gesturing to the conservative panelist] gets to hit you with it
anytime he wants. It doesn't cost money, it costs votes. It costs airtime and column inches.
You know why people don't like liberals? Because they lose. If liberals are so friggin’ smart,
how come they lose so GODDAM ALWAYS!
And [to the conservative panelist] with a straight face, you're going to tell students that
America's so starspangled awesome that we're the only ones in the world who have freedom?
Canada has freedom, Japan has freedom, the UK, France, Italy, Germany, Spain, Australia,
Belgium has freedom. Two hundred seven sovereign states in the world, like 180 of them
have freedom.
And you—sorority girl—yeah—just in case you accidentally wander into a voting booth one
day, there are some things you should know, and one of them is that there is absolutely no
evidence to support the statement that we're the greatest country in the world. We're seventh
in literacy, twenty-seventh in math, twenty-second in science, forty-ninth in life expectancy,
178th in infant mortality, third in median household income, number four in labor force, and
number four in exports. We lead the world in only three categories: number of incarcerated
citizens per capita, number of adults who believe angels are real, and defense spending,
where we spend more than the next twenty-six countries combined, twenty-five of whom are
allies. None of this is the fault of a 20-year-old college student, but you, nonetheless, are
without a doubt, a member of the WORST-period-GENERATION-period-EVER-period, so
when you ask what makes us the greatest country in the world, I don't know what the hell
you're talking about?! Yosemite?!!!
We sure used to be. We stood up for what was right! We fought for moral reasons, we passed
and struck down laws for moral reasons. We waged wars on poverty, not poor people. We
sacrificed, we cared about our neighbors, we put our money where our mouths were, and we
never beat our chest. We built great big things, made ungodly technological advances,
explored the universe, cured diseases, and cultivated the world's greatest artists and the
world's greatest economy. We reached for the stars, and we acted like men. We aspired to
intelligence; we didn't belittle it; it didn't make us feel inferior. We didn't identify ourselves by
who we voted for in the last election, and we didn't scare so easy. And we were able to be all
these things and do all these things because we were informed. By great men, men who were
revered. The first step in solving any problem is recognizing there is one—America is not the
greatest country in the world anymore.
`

func TestDeployError_Error(t *testing.T) {
	// mock our http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, expected)
		if err != nil {
			return
		}
	}))
	defer svr.Close()

	// resources for making our little http client
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		req, _ := http.NewRequestWithContext(ctx, "GET", svr.URL, nil)
		resp, _ := http.DefaultClient.Do(req)
		errChan <- &deployError{response: resp}
	}()

	// collect the error from the client
	capturedErr := <-errChan

	// simulate a cleanup of resources, the client is now done with the request so cancel the ctx
	cancel()

	assert.Error(t, capturedErr)
	assert.Equal(t, expected, capturedErr.Error())
}
