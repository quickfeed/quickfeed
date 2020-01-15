package main.java.autograder;

//Score is an object used to encode/decode a score from a test or tests. When a
//test is passed or a calculation of partial passed test is found, output a
//JSON object representing this object.
//
//Secret read from the output steam need to correspond to the course identifier
//given on the teachers panel. All other output will be ignored.
//
//With the formula in the Autograder CI the score percentage is calculated
//automatically. Give any max score, then pass on a given score the student
//gets for passed sub test within this the max score. Finally, set a weight
//it should have on the total. The weight does not need to within 100 or a
//percentage. If you want to only give a score for completing a test, then
//MaxScore == Score.
//
//Calculations in the CI follows this formula:
//total_weight    = sum(Weight)
//task_score[0:n] = Score[i] / MaxScore[i], gives {0 < task_score < 1}
//student_score   = sum( task_score[i] * (Weight[i]/total_weight) ), gives {0 < student_score < 1}
public class Score {
	private static String GlobalSecret;
	private String Secret;
	private String TestName;
	private int Score;
	private int MaxScore;
	private int Weight;
	
	// GolbalSecret represents the unique course identifier that will be used in
	// the Score struct constructors. Users of this package should set this
	// variable appropriately before using any exported
	// function in this package.
	public static void SetGlobalSecret(String secret) {
		GlobalSecret = secret;
	}
	
	public Score(String testname, int weight, int maxscore, int initialscore, String secret) {
		this.TestName = testname;
		this.Weight = weight;
		this.MaxScore = maxscore;
		this.Score = initialscore;
		this.Secret = secret;
	}
	
	public Score(String testname, int weight, int maxscore, int initialscore) {
		this.TestName = testname;
		this.Weight = weight;
		this.MaxScore = maxscore;
		this.Score = initialscore;
		this.Secret = this.GlobalSecret;
	}
	
	public Score(String testname, int weight, int maxscore) {
		this.TestName = testname;
		this.Weight = weight;
		this.MaxScore = maxscore;
		this.Secret = this.GlobalSecret;
	}
	
	// Inc will increase the score with one. 
	public void Inc(){
		if(this.Score < this.MaxScore){
			this.Score++;
		}
	}
	
	// Dec will decrease the score with one.
	public void Dec(){
		if(this.Score != 0){
			this.Score--;
		}
	}
	
	// IncBy will increase the score with the given amount.
	public void IncBy(int points){
		if(this.Score + points >= this.MaxScore){
			this.Score = this.MaxScore;
		} else {
			this.Score += points;
		}
	}
	
	// DecBy will decrease the score with given amount. 
	public void DecBy(int points){
		if(this.Score - points <= 0){
			this.Score = 0;
		} else {
			this.Score -= points;
		}
	}
	
	// PrintJSON will print JSON data representing this object to the output stream.  
	public void PrintJSON(){
		System.out.println(this.toJSON());
	}
	
	// toJSON will convert this object to JSON data.
	public String toJSON(){
		return "{"
					+ "\"Secret\":\"" + this.Secret +"\","
					+ "\"TestName\":\"" + this.TestName + "\","
					+ "\"Score\":" + this.Score + ","
					+ "\"MaxScore\":" + this.MaxScore + ","
					+ "\"Weight\":" + this.Weight
				+ "}";
	}
	
	// Print will print a summary of the score to the output stream.
	public void Print(){
		System.out.println(this.toString());
	}
	
	// toString will create a string with a summary of the score. 
	public String toString() {
		return this.TestName + ": " + this.Score + "/" + this.MaxScore + " cases passed.";
	}
}

