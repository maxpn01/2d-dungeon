package worldmap

const forestVertex = `#version 330
in vec3 vertexPosition;
in vec3 vertexNormal;
in vec4 vertexColor;
in vec2 vertexTexCoord;
uniform mat4 mvp;
uniform mat4 matModel;
uniform float mode;
uniform float clock;
out vec3 worldPos;
out vec3 normal;
out vec4 tint;
out vec2 uv;
void main() {
    vec3 p = vertexPosition;
    if (int(mode + 0.5) == 0 && p.y > 40.0 && vertexColor.a > 0.998) {
        float sway = sin(clock*1.3 + p.x*0.013 + p.z*0.017) * (p.y - 40.0) * 0.035;
        p.x += sway;
        p.z += sway*0.4;
    }
    if (int(mode + 0.5) == 2) {
        p.y += sin(clock*1.6 + p.x*0.05 + p.z*0.04) * 1.2;
    }
    worldPos = (matModel*vec4(p, 1.0)).xyz;
    normal = normalize(mat3(matModel)*vertexNormal);
    tint = vertexColor;
    uv = vertexTexCoord;
    gl_Position = mvp*vec4(p, 1.0);
}`

const forestFragment = `#version 330
in vec3 worldPos;
in vec3 normal;
in vec4 tint;
in vec2 uv;
uniform float mode;
uniform float clock;
uniform vec3 sunDir;
uniform vec3 sunColor;
uniform vec3 ambient;
uniform vec3 lantern;
uniform vec3 cameraPos;
uniform vec2 playerScreen;
uniform float playerDepthZ;
uniform sampler2D texture0;
uniform sampler2D texture1;
uniform vec4 colDiffuse;
uniform vec4 fogArea;
uniform vec3 hazeColor;
out vec4 finalColor;
float hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
float noise(vec2 p) {
    vec2 i = floor(p), f = fract(p);
    f = f*f*(3.0-2.0*f);
    return mix(mix(hash(i), hash(i+vec2(1,0)), f.x), mix(hash(i+vec2(0,1)), hash(i+vec2(1,1)), f.x), f.y);
}
float fogAt(vec2 p) {
    vec2 t = (p - fogArea.xy) / fogArea.zw;
    if (t.x < 0.0 || t.y < 0.0 || t.x > 1.0 || t.y > 1.0) return 1.0;
    return texture(texture1, t).r;
}
vec3 light(vec3 base, vec3 n) {
    float sun = max(dot(n, normalize(-sunDir)), 0.0);
    float sky = 0.55 + 0.45*n.y;
    vec3 c = base*(ambient*sky + sunColor*sun);
    float d = length(worldPos.xz - lantern.xz);
    float glow = clamp(1.0 - d/lantern.y, 0.0, 1.0);
    c += base*vec3(1.0, 0.78, 0.5)*glow*glow*0.85;
    return c;
}
vec3 atmosphere(vec3 c) {
    float fog = fogAt(worldPos.xz);
    float mist = noise(worldPos.xz*0.006 + vec2(clock*0.03, -clock*0.02))*0.6 + noise(worldPos.xz*0.015 - clock*0.02)*0.4;
    vec3 fogTint = mix(vec3(0.035, 0.05, 0.085), vec3(0.075, 0.09, 0.14), mist);
    float far = clamp((length(worldPos - cameraPos) - 1500.0)/2600.0, 0.0, 1.0);
    c = mix(c, hazeColor, far*0.55);
    return mix(c, fogTint, fog);
}
void main() {
    if (int(mode + 0.5) == 3) {
        vec4 t = texture(texture0, uv);
        float fog = fogAt(worldPos.xz);
        finalColor = vec4(t.rgb*(1.0 - fog*0.9), t.a);
        return;
    }
    if (int(mode + 0.5) == 1) {
        finalColor = vec4(0.0, 0.02, 0.04, tint.a*(1.0 - fogAt(worldPos.xz)*0.7));
        return;
    }
    vec3 base = tint.rgb*colDiffuse.rgb;
    if (int(mode + 0.5) == 2) {
        vec2 p = worldPos.xz;
        float w = noise(p*0.03 + vec2(clock*0.25, clock*0.1))*0.5 + noise(p*0.07 - vec2(clock*0.18, -clock*0.22))*0.5;
        vec3 deep = vec3(0.05, 0.16, 0.24), shallow = vec3(0.12, 0.38, 0.46);
        vec3 c = mix(deep, shallow, w);
        float sparkle = pow(noise(p*0.25 + clock*0.6), 12.0)*2.5;
        c += vec3(0.75, 0.9, 1.0)*sparkle;
        c = light(c, vec3(0.0, 1.0, 0.0)) + c*0.35;
        finalColor = vec4(atmosphere(c), 0.92);
        return;
    }
    if (int(mode + 0.5) == 4) {
        vec2 p = worldPos.xz;
        float g = noise(p*0.045)*0.6 + noise(p*0.16)*0.4;
        base *= 0.86 + g*0.28;
        float speck = step(0.93, hash(floor(p/4.0)));
        base = mix(base, base*1.25 + vec3(0.03, 0.04, 0.0), speck*0.5);
    }
    float emissive = clamp((0.97 - tint.a)/0.75, 0.0, 1.0);
    vec3 c = light(base, normalize(normal));
    if (int(mode + 0.5) == 0) {
        float ao = mix(0.55, 1.0, clamp(worldPos.y/45.0, 0.0, 1.0));
        c *= ao;
        if (tint.a > 0.998 && worldPos.y > 18.0 && worldPos.z > playerDepthZ + 10.0) {
            float d = length(gl_FragCoord.xy - playerScreen);
            if (d < 95.0) {
                float k = clamp((95.0 - d)/40.0, 0.0, 1.0)*0.8;
                float pattern = fract((floor(gl_FragCoord.x) + floor(gl_FragCoord.y)*2.0)*0.25 + floor(gl_FragCoord.y)*0.5);
                if (pattern < k) discard;
            }
        }
    }
    c = mix(c, base*1.7, emissive);
    finalColor = vec4(atmosphere(c), 1.0);
}`
