package test;

import java.io.File;
import java.io.FileNotFoundException;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class Interpreter {
    public Interpreter() {

    }

    public static void test() {
        for (byte i = 0; i < 128; i++) {
            System.out.println(1);
        }
    }

    public static void main(String[] args) throws FileNotFoundException {

//        ((Interpreter) null). test();
        Scanner scanner = new Scanner(new File("./1"));
        StringBuilder sb = new StringBuilder();
        while (scanner.hasNext()) sb.append(scanner.nextLine()).append('\n');
        Interpreter interpreter = new Interpreter();
        System.out.println(interpreter.parseCode(sb.toString()).interpreter());
        System.out.println(interpreter.calculator("3+33*((5-6)+.72-18.)/23.14"));
    }

    private Map<String, Double> map = new HashMap<>();
    private Queue<String[]> codes = new ArrayDeque<>();

    private Interpreter parseCode(String src) {
        for (String code : src.replaceFirst("return ", "#=").split("\n")) {

            String[] ES = code.split("=");
            map.put(ES[0], null);
            codes.offer(ES);
            if (code.startsWith("#=")) break;
        }
        return this;
    }

    private double interpreter() {
        codes.forEach(exp -> map.put(exp[0], calculator(exp[1])));
        return map.get("#");
    }

    private double calculator(String code) {
        Matcher matcher = Pattern.compile("(\\d*+\\.\\d*+)|(\\d++)|([a-zA-Z]++)|(.)").matcher(code + '#');
        Stack<Character> operations = new Stack<>();
        Stack<Double> operands = new Stack<>();
        operations.push('#');
        while (matcher.find()) {
            String decimal = matcher.group(1);
            String integer = matcher.group(2);
            String variable = matcher.group(3);
            String operation = matcher.group(4);
            if (decimal != null) {
                operands.push(Double.valueOf(decimal));
            } else if (integer != null) {
                operands.push(Double.valueOf(integer));
            } else if (variable != null) {
                Double value = map.get(variable);
                if (value == null)
                    throw new NullPointerException("in " + code + ", variable " + variable + " has not been initialized");
                operands.push(value);
            } else if (operation != null) {
                Character left = operations.pop();
                char right = operation.charAt(0);
                while (compare(left, right) >= 0) {
                    System.out.println(operands);
                    System.out.println(operations);
                    Double R = operands.pop();
                    Double L = operands.pop();
                    operands.push(cal(L, left, R));
                    left = operations.pop();
                    if (right == ')' && left == '(') {
                        break;
                    }
                    if (right == '#' && left == '#') {
                        return operands.peek();
                    }
                }
                if (right != ')') {
                    operations.push(left);
                    operations.push(right);
                }
            }
        }
        throw new RuntimeException();
    }

    private double cal(Double a, Character operation, Double b) {
        System.out.println(a + ", " + operation + ", " + b);
        switch (operation) {
            case '+':
                return a + b;
            case '-':
                return a - b;
            case '*':
                return a * b;
            case '/':
                return a / b;
            default:
                throw new IllegalArgumentException("undefined operation: " + operation);
        }
    }

    public int compare(char left, char right) {
        if (left == '(' || left == '#') {
            return -1;
        }
        if (right == ')' || right == '#') {
            return 1;
        }
        char[][] ops = {
                {'#',},
                {'+', '-'},
                {'*', '/'},
                {'('}
        };
        Integer p = null;
        Integer q = null;
        for (int i = ops.length - 1; i >= 0; i--) {
            for (int j = ops[i].length - 1; j >= 0; j--) {
                if (left == ops[i][j]) {
                    p = i;
                }
                if (right == ops[i][j]) {

                    q = i;
                }
                if (p != null && q != null) {
                    return p - q;
                }
            }
        }
        throw new IllegalArgumentException(left + ", " + right);
    }
}
